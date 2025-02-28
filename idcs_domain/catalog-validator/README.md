# About
Catalog Validator is a CLI tool for Product Catalog Developers to quickly validate new custom resources like vendors, products, etc before adding them to product catalog.

## Prerequisites

1. Install go.
2. You will need a source folder where required Custom Resource Definitions are stored as yaml files.

3. Run:

    3.1 Using Makefile
    
    ```  
    make install
    ```

    3.2 Using go
    
    ```
    go install
    ```

## Run CLI commands to validate resources

### Usage
```
catalog-validator [command] [args] [flags]
```

### Samples
- catalog-validator about
- catalog-validator help
- catalog-validator validate --help
- catalog-validator validate file1.yaml file2.yaml --src=path/to/CRD
- catalog-validator validate file1.yaml file2.yaml -s=path/to/CRD
