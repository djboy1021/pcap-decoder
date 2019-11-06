# PCAP Decoder
This tool extracts the 3D points stored in a PCAP file.
A 3D point contains the following information
* **Distance** *(from the lidar)*
* **Azimuth**
* **X** coordinate
* **Y** coordinate
* **Z** coordinate
* **LaserID**
* **Intensity**
* **Timestamp**

A JSON file is currently set as the output file

## Command line API
The following are the acceptable input arguments. They can be written in any order
* **--pcapFile**
  * All characters between *--pcapFile* and the next *--\<key\>* will be interpreted as the input PCAP file
  * The path string may not be enclosed with a quoutation mark.
  * Paths that contain spaces are properly handled
* **--outputPath**
  * All characters between *--outputPath* and the next *--\<key\>* will be interpreted as the location of the output files
* **--mkdirp**
  * This accepts a boolean *(**true** or **false**)* as input
  * When this is set to true, the output path will be created recursively
* **--startFrame**
  * This accepts a positive integer as input
  * **startFrame** together with the **endFrame** can be used to selectively parse particular frames from a PCAP file.
  * default value is 0.
* **--endFrame**
  * This accepts an integer as input
  * default value is -1 or the end frame

### Example


```console
$ ./pcapDecoder.exe --pcapFile "C:/Users/brendon.dulam/Desktop/Magic Hat/city.pcap" --outputPath 'V:/JP01/DataLake/Common_Write/CLARITY_OUPUT/Magic_Hat/json/test' --startFrame 0 --endFrame 20 --mkdirp false
```

## Supported Models
The following Velodyne Lidar models are currently supported
* VLP32
* VLP16

