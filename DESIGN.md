## Storing Data
This project stores data in good old tables in Potsgres. I chose Postgres because it is a reliable, widely-used relational database that offers ACID, 
which is great for handling transactions, but the need for strong consistency depends on the business requirements i.e. if we strictly need all transactions to be consistent.
I could go for something more scalable that allows partitioning if we're ok with eventual consistency.
The schema is designed to optimize performance, with appropriate indexing to speed up queries, especially for the offers eligibility endpoint. I used int to store money values as suggested, but it's also worth thinking about numeric type, as it supports more decimals and it'd be better if multiple currencies are 
a possibility.
I decided to use `golang migrate` to manage the migrations and they are located in `internal/repository/db/migrations`. There are 3 tables: merchants, transactions and offers and they have foreign key constraints where they relate.
There's no endpoint to create merchants, so they have to be seeded. 
This project uses go-jet to build the queries using idiomatic Go code. It has better performance than an ORM and I think it offers a nice flexibility to write SQL queries as 
close to actual SQL as possible. Of course an ORM could speed up development a bit, so this is trade off to notice.
## Eligibility Logic
I calculated the eligibility using a single SQL query located in `internal/repository/db/offers.go`. The query is becoming a bit complex, so a discussion could be made
on moving the logic to the core layer and making simpler queries. While in a scalable architecture this service would be running in pods, which would allow dividing
the load, Postgres is optimized for doing the query, so if it ever were to become a bottleneck, exploring and experimenting would be a good idea. 
The query applies all dates, MCCs and merchant filters using JOINS and subqueries. An alternative to explore in the future would be pre processing the eligibility and storing it 
ready to be consumed by the endpoint instead of calculating it everytime, even though this would affect consistency.
## Code architecture
It's worth mentioning that this is a small service for now, so the packages structure wasn't absolutely necessary, but I always like to keep code more organized, not to
mention this increases the flexibility and maintainability of the code as the service grows. You spend more time in the beginning, but it pays off as it grows.
## Considerations
This service is missing a few things in the interest of time. Adding an auth middleware for stateless authentication would be needed for a prod env. The service
is not validating any of the inputs currently, so no input errors are returned. I'd like to add created_at and updated_at fields to the tables for the sake of auditing if ever needed.
It would be necessary to add observability to it as well, using open telemetry. Depending on the use case of the eligibility offers endpoint, I'd consider 
paginating the results. I also decided not to upsert offers (only create) to make things simpler, but it shouldn't be too hard to add that logic. I've created the standard error message
in the API, but creating more errors to match with different status codes would be better. At last, I think a swagger generating tool would also be really important if this service
were to be used externally.
