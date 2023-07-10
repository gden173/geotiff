# GeoTIFF

<!--toc:start-->
- [Geotiff](#geotiff)
- [Description](#description)
- [Specification](#specification)
- [Test Data](#test-data)
<!--toc:end-->


# Description 

This package contains a minimal implementation of a GeoTIFF reader. Currently
it only has the capabilities to parse GeoTIFF in a tile layout containing a
double float data type.

Only a subset of the TIFF and GeoTIFF tags are implemented for this particulars
use case.


# Specification 

The specifications which where used to implement this GeoTIFF are listed below

- [Adobes TIFF 6.0 Specification](https://web.archive.org/web/20180810205359/https://www.adobe.io/content/udp/en/open/standards/TIFF/_jcr_content/contentbody/download/file.res/TIFF6.pdf)
- [OGC GeoTIFF Standard](https://web.archive.org/web/20220120213932/https://docs.opengeospatial.org/is/19-008r4/19-008r4.pdf)

A useful resource for other GeoTIFF tag codes and specifications is also 

- [GeoTIFF MapTools](https://web.archive.org/web/20220308133121/http://geotiff.maptools.org/spec/geotiff6.html)

# Test Data 

For testing GeoTIFF parsing test data is placed in the `testdata` folder. 

That test data was sourced from 

- [DEM_SRTM_1Second_Hydro_Enforced](https://web.archive.org/web/20230330031117/https://services.ga.gov.au/site_9/services/DEM_SRTM_1Second_Hydro_Enforced/MapServer/WCSServer?request=GetCapabilities&service=WCS&f=geotiff)[CC 4.0](https://creativecommons.org/licenses/by/4.0/legalcode)


