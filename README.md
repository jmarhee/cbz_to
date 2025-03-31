# cbz_to

CLI to convert .cbz files to .epub,.pdf, or .mobi.

## Usage

Convert a file using the list of files as arguments:

```bash
cbz_to --format=$OUTPUT_FORMAT file1.cbz...fileX.cbz
```

The following output formats are supported:
* epub
* mobi
* pdf

## Testing

Includes a functional test that creates a dummy .cbz file and tests the two preferred formats (.epub and .pdf):

```bash
go test
```

Successful tests will read:

```
Converted test.cbz to test.pdf
Converted test.cbz to test.epub
PASS
ok  	personal	0.426s
```
