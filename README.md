
# LLM Context Generator

`llm-context-generator` is a Go program designed to scan a directory recursively for files matching specified file extension patterns, and concatenate their contents into a single text file. This makes it easy to feed relevant files from a project into a large language model (LLM) context.

## Features

- Recursively scans a specified source directory.
- Filters files based on provided file extension patterns.
- Concatenates the contents of matched files onto the clipboard or into an output file.
- Easy to use with command line arguments.

## Installation
1. Ensure you have Go installed on your machine. You can download it from the [official website](https://golang.org/dl/).
2. Clone the repository or download the `llm-context-generator` source code.
3. Build the program:

```sh
go build -o context_generator context_generator.go
```

## Usage
Run the context_generator with the required argument only, filename patterns to match. This writes to the clipboard by default, using the current directory.

```sh
./context_generator -p <patterns>
```

Or output to the console:

```sh
./context_generator -p <patterns> -c
```

Alternatively write to an output file, and specify the source directory as well as patterns to exclude:

```sh
./context_generator -s <source directory> -o <output file> -p <patterns> -e <excludes>
```

### Arguments

-s: Specifies the source directory to scan.

-o: Specifies the path to the output file where the concatenated contents will be stored.

-c: Output the content to the console.

-p: Specifies the file extension patterns to match, separated by semicolons (;).

-e: Specifies the file path patterns to exclude, separated by semicolons (;).

### Example

```sh
./context_generator -s /path/to/source -o /path/to/output.txt -p "java;yaml;kts" -e "test;example"
```
  
In this example:

* The program will scan /path/to/source directory recursively.
* It will match files with extensions .java, .yaml, and .kts.
* It will exclude files with paths containing test or example.
* The contents of matched files will be concatenated into /path/to/output.txt.
 

### Output Format

The contents of each matched file will be wrapped with markers indicating the filename, like so:
  
```
=filename=
<file  contents>
=/filename=
=filename 2=
<file  2  contents>
=/filename 2=
```
