### What is Aerospike?
* Fast and Scaleable Key/Value Store.
* *NOT* a general purpose database.

### Why Go?
* **Fast**: 
	* As fast or faster than Synchronous Java depending on type of operation, and an order of magnitude faster AND simpler than Java


* **Simple, Productive**:
	* Less abstractions to work with
	* Less boiler-plate helps readability and maitainability
	* Less patterns, more composition

	
* **Concurrent, Non-blocking**:
	* Concurrency is key to scaling up
	* Slow nodes/connections won't block the whole client
	* Simple and understandable patterns

* Generally **fun** to work with - minimalism is art, and as enjoyable

### Aerospike Terminalogy
* **Cluster**: A collection of preconfigured database nodes
* **Namespace**: Database (needs setup in config file)
* **Set**: Table
* **Record**: Record, Row; In Aerospike, each record has a Generation and an Expiration time, which are accessible to users
* **Bin**: Column
* **UDF**: User Defined Functions
* **Key**: Primary key; It embeds Namespace and Set besides the key's value

### Aerospike Terminalogy - Cont'd
* **Put**: UPDATE OR INSERT ... where ...
* **Get**: SELECT ... FROM NAMESPACE.SET WHERE KEY = ```key```
* **Delete**: DELETE FROM NAMESPACE.SET where KEY = ```key```
* **Exists**: *true* if a record with provided key exists, *false* otherwise
* **Touch**: updates the record metadata with new values
* **BatchGet**, **BatchExists**: Like *Get* and *Exists*, but for more than one key, in one request to reduce network io.
* **Scan**: SELECT ... FROM NAMESPACE.SET
* **Query**: SELECT ... FROM NAMESPACE.SET WHERE INDEXED_BIN [== ```value``` | BETWEEN ```value1``` AND ```value2```]


### Client Data Model
* ```Key```: Encapsulates Namespace, Set and key value.
* ```Bin```: Each bin is like a column, and the name and value are encapsulated in the Bin object.
* ```Record```: Anytime you ask for a predetermined number of records from the database, the result will return encapsuaed in a ```Record``` or ```[]Record```
* ```Client```: Client encapsulates connection to an Aerospike cluster. All cluster maintenance, connection and buffer pooling and other such concurrent voodoo is managed automatically inside.
* ```Recordset```: Encapsulates the results of ```Scan``` and ```Query``` operations. It has two channels which return ```Record```s and ```error```s.

### Client Data Model - Policies
Different policies affect how client works, or records are affected by commands.

* ```BasePolicy```: Encapsulates variables like the timeout and priority of the operation.
* ```WritePolicy```: Embeds ```BasePolicy``` and adds Generation, Exiration (TTL), and policies affecting the record.
	* ```RecordExistsAction```: Specifies what to do when the record already exists.
	* ```GenerationPolicy```: Specifies what generation to expect before writing.
* ```ScanPolicy```: Percent of data scanned, serial or concurrent scan, record queue size, ....
* ```QueryPolicy```: Record queue size, block until migrations are over, ....

### Client API - Common Operations

* ```Get()```, ```Put()```, ```Delete()```
* ```Touch()```, ```Exists()```
* ```GetHeader()```, ```BatchGetHeader()```
* ```BatchGet()```, ```BatchExists()```
* ```Append()```, ```Prepend()```, ```Add()```

### Client API - Record-level Atomic Multiple Operations

* ```Operate()```: Can affect an operation on multiple Bins in a record (Put, Add, Prepend, ...) in a single atomic request.

### Client API - UDF

* ```RegisterUDF()```: Registers a UDF on the server
* ```RemoveUDF()```: Drop a UDF
* ```ListUDF()```: List all UDFs registered on the server
* ```Execute()```: Executes a UDF on server and returns the result


### Advanced Pattern Example: Cluster-wide Safe Locking

Problem Description: Want to set a lock between all clients on the cluster.

1. How to do it without race-conditions?	
2. How to avoid dead-lock and safely unlock, even if the client fails to do so?

Solution: If we could create a record to represent a lock with the following conditions:
1. Only one record should be created
2. Should be created only for one client, and fail for others

### Solution - Part 1
```go

writePolicy := NewWritePolicy(0, 0)
writePolicy.RecordExistsAction = CREATE_ONLY

// Will pass for only one request, and will fail for others
err := client.Put(writePolicy, key, bins...)
```

This doesn't address how to safely unlock. What if the client crashes and fails to unlock?

### Solution - Part 2
```go

// set a TTL; will automatically expire and the record will be deleted
writePolicy := NewWritePolicy(0, ```10```)
writePolicy.RecordExistsAction = CREATE_ONLY

// Will pass for only one request, and will fail for others
err := client.Put(writePolicy, key, bins...)
```

AVOID MISUSE - DISTRIBUTED COORDINATED LOCKING AND SYNCHRONIZATION IS NON-TRIVIAL
