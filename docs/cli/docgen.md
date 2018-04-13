# CLI doc generator  
Generates CLI docs for `latticectl` and `laasctl`.

## Usage
### Input
The command struct is already included in the binary.  

You can also attach extra markdown documentation to each command. 
It needs to be placed in the directory specified by `--input-docs` and then under the directory tree corresponding to the command hierarchy.  
For example, if you want to add extra description to the `latticectl` docs for the `latticectl services addresses` command, 
you'd put a `description.md` file like so:
`<latticectl-root-docs-dir>/services/addresses/description.md`.  

### Output
All docs for a given repo will be output into a single `docs.md` file in the directory specified by `--output-docs`.

**Flags**

| Name | Description |  
| --- | --- |  
|`--input-docs INPUT-DOCS` | Extra markdown docs input directory (default value: `./docs/cli`) |  
|`--output-docs OUTPUT-DOCS` | Markdown docs output file path (default value: `./doc.md`) |