package gentags

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/perrotbryan/gentags/templates"

	"golang.org/x/tools/go/packages"
)

type TaggedField struct {
	FieldName string
	FieldType string
	TagName   string
	TagValue  string
	IsPointer bool
	IsNumeric bool
	IsBasic   bool
	Options   []string
}

type FieldTypeInfo struct {
	TypeString string
	IsPointer  bool
	IsNumeric  bool
	IsBasic    bool
}

type TaggedStruct struct {
	Name   string
	Fields []TaggedField
}

type TaggedFile struct {
	Name    string
	Structs []TaggedStruct
}

type TaggedPackage struct {
	Name  string
	Files []TaggedFile
}

type TagEntry struct {
	Packages []TaggedPackage
	Template *template.Template
}

type TagMap map[string]TagEntry

type Generator struct {
	typesInfo  *types.Info
	currentPkg *types.Package
}

func (generator *Generator) Analyze(dir string) []TaggedPackage {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedModule,
		Dir:   dir,
		Fset:  token.NewFileSet(),
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Fatal("package loading errors occurred")
	}

	var result []TaggedPackage

	for _, pkg := range pkgs {
		generator.typesInfo = pkg.TypesInfo
		generator.currentPkg = pkg.Types

		taggedPkg := TaggedPackage{
			Name:  pkg.Name,
			Files: []TaggedFile{},
		}

		for i, file := range pkg.Syntax {
			// skip generated files
			if ast.IsGenerated(file) {
				continue
			}

			fileName := pkg.CompiledGoFiles[i]
			taggedFile := generator.parseFile(fileName, file)
			if len(taggedFile.Structs) > 0 {
				taggedPkg.Files = append(taggedPkg.Files, taggedFile)
			}
		}

		if len(taggedPkg.Files) > 0 {
			result = append(result, taggedPkg)
		}
	}

	return result
}

func (generator *Generator) parseFile(fileName string, file *ast.File) TaggedFile {
	// We cannot pre-allocate enough space, as we loop on Decls, which is all declarations,
	// not just structs. Since there can be a metric ton of declarations, it's better
	// to resize as we need
	structs := []TaggedStruct{}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		// Ignore non-type declaration
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			// Ignore non-struct declaration
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Here, just like for structs, we're unable to pre-allocate capacity
			// since the resulting fields can be > to Fields.List length
			// in the case of multiple fields in a single line
			fields := []TaggedField{}
			hasTaggedFields := false

			for _, field := range structType.Fields.List {
				if field.Names == nil {
					// Embedded field #TODO
					continue
				}

				// No tag, skip
				if field.Tag == nil {
					continue
				}

				hasTaggedFields = true

				typeInfos := generator.extractFieldTypeInfo(field.Type)

				// Go allows multiple declarations in a single line eg:
				// foo, bar string `snafu:"baz"``
				// So we loop on each and treat them as different fields
				// Note that this isn't compatible with some tags, such as json, that require
				// tags unicity
				for _, name := range field.Names {
					rowFields := parseStructTag(name.Name, field.Tag.Value, typeInfos)
					fields = append(fields, rowFields...)
				}
			}
			if hasTaggedFields {
				structs = append(structs, TaggedStruct{Name: typeSpec.Name.Name, Fields: fields})
			}
		}
	}

	return TaggedFile{Name: fileName, Structs: structs}
}

func parseStructTag(fieldName string, fieldTag string, typeInfos FieldTypeInfo) []TaggedField {
	tag := reflect.StructTag(strings.Trim(fieldTag, "`"))
	fields := []TaggedField{}

	for tagStr := range strings.SplitSeq(string(tag), " ") {
		keySep := strings.Index(tagStr, ":")
		if keySep <= 0 {
			continue
		}

		key := tagStr[:keySep]
		value, ok := tag.Lookup(key)
		if !ok {
			continue
		}

		parts := strings.Split(value, ",")
		fields = append(fields, TaggedField{
			FieldName: fieldName,
			TagName:   key,
			TagValue:  parts[0],
			FieldType: typeInfos.TypeString,
			IsPointer: typeInfos.IsPointer,
			IsNumeric: typeInfos.IsNumeric,
			IsBasic:   typeInfos.IsBasic,
			Options:   parts[1:],
		})
	}

	return fields
}

func mapByTagName(pkgs []TaggedPackage) TagMap {
	result := make(TagMap)

	for _, pkg := range pkgs {
		tagToPkg := make(map[string]TaggedPackage)

		for _, file := range pkg.Files {
			tagToFile := make(map[string]TaggedFile)

			for _, strct := range file.Structs {
				tagToStruct := make(map[string]TaggedStruct)

				for _, field := range strct.Fields {
					// Initialize TaggedStruct in its tag's group
					ts, ok := tagToStruct[field.TagName]
					if !ok {
						ts = TaggedStruct{
							Name:   strct.Name,
							Fields: []TaggedField{},
						}
					}
					ts.Fields = append(ts.Fields, field)
					tagToStruct[field.TagName] = ts
				}

				// Push structs into the right file for each tag
				for tag, ts := range tagToStruct {
					tf, ok := tagToFile[tag]
					if !ok {
						tf = TaggedFile{
							Name:    file.Name,
							Structs: []TaggedStruct{},
						}
					}
					tf.Structs = append(tf.Structs, ts)
					tagToFile[tag] = tf
				}
			}

			// Push files into the right package for each tag
			for tag, tf := range tagToFile {
				tp, ok := tagToPkg[tag]
				if !ok {
					tp = TaggedPackage{
						Name:  pkg.Name,
						Files: []TaggedFile{},
					}
				}
				tp.Files = append(tp.Files, tf)
				tagToPkg[tag] = tp
			}
		}

		// Add each per-tag package to the result
		for tag, tp := range tagToPkg {
			te, ok := result[tag]
			if !ok {
				var tplt *template.Template
				tplt, err := templates.LoadTemplate(tag)
				if err != nil {
					fmt.Println(fmt.Errorf("LoadTemplate error: %w", err))
				}

				te = TagEntry{
					Packages: []TaggedPackage{},
					Template: tplt,
				}
				result[tag] = te
			}
			te.Packages = append(te.Packages, tp)
			result[tag] = te
		}
	}

	return result
}

func (a *Generator) Generate(pkgs []TaggedPackage) error {
	tm := mapByTagName(pkgs)

	type genFileInfo struct {
		PackageName string
		Structs     []TaggedStruct
	}

	for tag, entry := range tm {
		if entry.Template == nil {
			continue
		}

		for _, pkg := range entry.Packages {
			for _, file := range pkg.Files {
				gfi := genFileInfo{
					PackageName: pkg.Name,
					Structs:     file.Structs,
				}
				tagFilePath := strings.TrimSuffix(
					filepath.Base(file.Name),
					filepath.Ext(file.Name),
				)
				tagFilePath += fmt.Sprintf("_%s.go", tag)

				err := func() error {
					var buf bytes.Buffer
					if err := entry.Template.ExecuteTemplate(&buf, "main", gfi); err != nil {
						return fmt.Errorf("%s template execute error: %w", tag, err)
					}

					formatted, err := format.Source(buf.Bytes())
					if err != nil {
						return fmt.Errorf("format error: %w\nRAW:\n%s", err, buf.String())
					}

					outFile, err := os.OpenFile(tagFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						return fmt.Errorf("create file %s error: %w", tagFilePath, err)
					}
					defer outFile.Close()

					_, err = outFile.Write(formatted)
					return nil
				}()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (generator *Generator) extractFieldTypeInfo(expr ast.Expr) FieldTypeInfo {
	info := FieldTypeInfo{}

	if generator.typesInfo == nil {
		return info
	}

	originalType := generator.typesInfo.TypeOf(expr)
	if originalType == nil {
		return info
	}

	t := originalType

	// Unwrap all pointer layers (handle **types)
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			break
		}
		info.IsPointer = true
		t = ptr.Elem()
	}

	// Resolve named types
	if named, ok := t.(*types.Named); ok {
		t = named.Underlying()
	}

	// Detect basic and numeric types
	if basic, ok := t.(*types.Basic); ok {
		info.IsBasic = true
		info.IsNumeric = basic.Info()&types.IsNumeric != 0
	}

	info.TypeString = types.TypeString(
		originalType,
		generator.qualifier,
	)

	return info
}

func (generator *Generator) qualifier(p *types.Package) string {
	if p == nil {
		return ""
	}

	// Do not qualify types from the same package
	if generator.currentPkg != nil && p.Path() == generator.currentPkg.Path() {
		return ""
	}

	return p.Name()
}
