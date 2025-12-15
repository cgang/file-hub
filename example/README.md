# File Hub Example

This directory contains example code and resources for the File Hub project.

## Contents

- `example_db_usage.go` - Example code demonstrating how to use the database access layer
- `root/` - Example root directory where files would be stored (in a real implementation)
- `propfind-allprop.xml` - Example PROPFIND request with allprop (returns all properties)
- `propfind-propname.xml` - Example PROPFIND request with propname (returns property names only)
- `propfind-prop.xml` - Example PROPFIND request with specific properties
- `example.xml` - Example PROPFIND response

## How to Run Example

To run the example, you'll first need to set up a PostgreSQL database and update the database connection parameters in the example code.

## WebDAV PROPFIND Requests

The File Hub supports three types of PROPFIND requests:

1. **allprop**: Returns all known properties for resources
2. **propname**: Returns only the names of properties available for resources
3. **prop**: Returns specific properties as requested

To test these with curl:

```bash
# allprop request
curl -X PROPFIND -H "Depth: 1" -H "Content-Type: text/xml" --data-binary @propfind-allprop.xml http://localhost:8080/dav/repo/

# propname request
curl -X PROPFIND -H "Depth: 1" -H "Content-Type: text/xml" --data-binary @propfind-propname.xml http://localhost:8080/dav/repo/

# prop request for specific properties
curl -X PROPFIND -H "Depth: 1" -H "Content-Type: text/xml" --data-binary @propfind-prop.xml http://localhost:8080/dav/repo/
```