
# LLM Context Generator

`llm-context-generator` is a Go program designed to scan a directory recursively for files matching specified file extension patterns, and concatenate their contents into a single text file. This makes it easy to feed relevant files from a project into a large language model (LLM) context.

## Features

- Recursively scans a specified source directory.
- Filters files based on provided file extension patterns.
- Concatenates the contents of matched files into a single output file with a clear structure for each file's content.
- Easy to use with command line arguments.

## Installation
1. Ensure you have Go installed on your machine. You can download it from the [official website](https://golang.org/dl/).
2. Clone the repository or download the `llm-context-generator` source code.
3. Build the program:

```ssh
go build -o context_generator context_generator.go
```

## Usage
Run the context_generator with the required command line arguments:

```ssh
./context_generator -s <source directory> -o <output file> -p <patterns> -e <excludes>
```

### Arguments

-s or --source: Specifies the source directory to scan.

-o or --output: Specifies the path to the output file where the concatenated contents will be stored.

-p or --patterns: Specifies the file extension patterns to match, separated by semicolons (;).

-e or --excludes: Specifies the file path patterns to exclude, separated by semicolons (;).

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
===filename===

<file  contents>

===/filename===

===filename 2===

<file  2  contents>

===/filename 2===
```
