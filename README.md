# Postcode

## Overview
The `postcode` package is designed for fetching and parsing postal code data, specifically from the GeoNames geographical database (https://www.geonames.org/). 
It offers a straightforward interface to download postal code data for various countries and parse them into a structured Go data type.

## Features
- Download postal code data by country from the GeoNames database.
- Utilizes HTTP ETag caching to minimize data transfer.
- Parses the downloaded data into a structured format for easy use in Go applications.

## Installation
To use the `postcode` package in your Go project, simply execute the following command:

```bash
go get github.com/ngrash/postcode
```

## Usage

### Fetching Postal Code Data
To fetch postal code data for a specific country, use the `FetchCountry` function. This function also supports ETag caching to minimize unnecessary data transfers.

Example:
```go
package main

import "github.com/ngrash/postcode"

func main() {
    var previousEtag string
    entries, modified, newEtag, err := postcode.FetchCountry("US", previousEtag)
    if err != nil {
        // Handle error
    }
    if modified {
        // Process new entries
        // Save newEtag for future requests
    }
}
```

### Fields in Postal Code Entry
Each postal code entry (`Entry` type) is an array of 12 strings, representing different data fields:

- `CountryCode`
- `PostalCode`
- `PlaceName`
- `AdminName1`
- `AdminCode1`
- `AdminName2`
- `AdminCode2`
- `AdminName3`
- `AdminCode3`
- `Latitude`
- `Longitude`
- `Accuracy`

## Contributing
Contributions to the `postcode` package are welcome. Please feel free to submit pull requests or open issues for bugs, feature requests, license problems or documentation improvements.

## License
This project is licensed under the [MIT License](LICENSE).

The data downloaded from [GeoNames.org](http://geonames.org) is licensed under [Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/).
This includes the ZIP file in the test_data directory as well as derived snippets used in the tests.
Postal code databases for some countries may come with additional licenses. See [GeoName's readme.txt](https://download.geonames.org/export/zip/readme.txt) for details.
