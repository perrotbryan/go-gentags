# gentags

**gentags** is a high-performance Go code generator that analyzes struct tags and generates optimized, type-safe code. It is designed to eliminate runtime reflection by producing static implementations tailored to your data structures.

The primary use case is generating custom serialization logic — such as reflection-free JSON encoders and decoders — based on struct tags like `json`, `bson`, or any custom tag.

It was originally implemented because I realized that, when performing JSON serialization/deserialization, reflect ran multiple times on the same structs, over and over again, which considerably slowed down critical path execution, most notably in a data chain. 
Although they are very convenient for the developer, such approaches do not really make sense in a production environment — another example of ease-of-use superseding performance.

---

## Features

* **Reflection-Free Runtime**
  * Generates highly optimized code without using `reflect` for known types.

* **High Performance**
  * Automatically generates code based on struct tags.
  
* **Recursive Type Support**
  * Handles nested structs and pointers using generated code.

* **Multi-Tag Support**
  * Supports any struct tag (as long as there is a template implementation for it).

* **Template-Based Generation**
  * Easily extensible via customizable templates.


---

## Example

### Input

```go
type Test struct {
	Name   string `json:"name" bson:"name"`
	Bar    int    `json:"bar"`
	Baz    string `json:"baz"`
	Parent *Test  `json:"parent,omitempty"`
}
```

### Generate Code

```go
//go:generate gentags -dir=.
```

Run:

```sh
go generate ./...
```

### Output

A file such as `types_json.go` will be generated, containing optimized implementations:

```go
func (o *Test) MarshalJSON() ([]byte, error)
func (o *Test) UnmarshalJSON(data []byte) error
```

These implementations avoid reflection and provide high-performance serialization.

---

## How It Works

### 1. Package Analysis

`gentags` scans your project using:

```go
golang.org/x/tools/go/packages
```

This ensures accurate, module-aware type information.

### 2. Struct Tag Extraction

It identifies structs and extracts tagged fields such as:

```go
`json:"name,omitempty"`
```

### 3. Type Inspection

Using `go/types`, it determines:

* Whether a field is a pointer
* Whether it is numeric or basic
* Its exact type and package

### 4. Template-Based Generation

Each tag corresponds to a template (e.g., `json.tplt`) that defines the generated code.

### 5. Code Emission

Generated files:

* Are written next to their source files
* Are automatically formatted using `go/format`
* Contain optimized serialization logic

---

## Reflection-Free

For known types, `gentags` generates code that avoids reflection entirely.
The JSON template 

| Type              | Strategy                    |
| ----------------- | --------------------------- |
| Strings           | `strconv.Unquote`           |
| Integers          | `strconv.ParseInt`          |
| Unsigned Integers | `strconv.ParseUint`         |
| Floats            | `strconv.ParseFloat`        |
| Booleans          | `strconv.ParseBool`         |
| Nested Structs    | Recursive generated code    |
| Pointers          | Explicit null handling      |
| Unknown Types     | Fallback to `encoding/json` |

This approach delivers better performance than reflection-based approaches while remaining simple and extensible.

---

## Installation

### Install from GitHub

```sh
go install github.com/perrotbryan/gentags/cmd/gentags@latest
```

Ensure your `$GOPATH/bin` is in your `PATH`:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Usage

### Step 1: Add a Generate Directive

```go
//go:generate gentags -dir=.
```

### Step 2: Run the Generator

```sh
go generate ./...
```

Generated files will follow the naming convention:

```
<source>_<tag>.go
```

Example:

```
types_json.go
```

---

## Creating Custom Templates

You can define your own templates for additional tags.

Example:

```
templates/
└── xml.tplt
```

The generator will automatically use it for struct tags using the filename as tag key:

```go
type Example struct {
	ID int `xml:"id"`
}
```

## Requirements

* Go **1.22+** (recommended)
* Go modules enabled

---

## License

This project is released under the MIT License.

---

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests to improve features, templates, or performance.

---

## Roadmap

* [ ] Support for slices and maps in reflection-free decoding
* [ ] Additional templates for common formats (XML, YAML, BSON)
* [ ] Benchmark suite

---

## Author

**Bryan Perrot**
GitHub: https://github.com/perrotbryan
