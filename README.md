## How to execute this service
You can get a bunch of helpers by executing the command `make help` in the root. The easiest way to run this service is by executing
`make up`. It will start docker-compose with all the necessary dependencies and it will seed 2 merchants in the database, as there's not an
endpoint to add them. All you need is docker and docker compose installed. The IDs of the fake merchants are: a1b2c3d4-e5f6-7890-abcd-ef1234567890 and 
a1b2c3d4-e5f6-7890-abcd-ef1234567890.
## Storage
I selected Postgres as the storage. I like it and it has a great set of features all around. I also think the ACID constraints are important for the transactions.
To ensure a better performance, indices were added, specially for the offers eligibility endpoint. 
## Tests
To run the tests, you can execute `make test`. This command will execute all tests in the project. The databse layer uses testcontainers for real database communication
which means you need docker. You could also run `make pretty_tests` which are a bit more readable, but it needs a tool described in the Makefile.
