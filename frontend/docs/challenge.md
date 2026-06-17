# Fullstack Challenge

Overview
This pet store will let merchants set up their store with listings for the pets they have to sell. Customers will be
able to login to the store and place orders for pets. For the purposes of reducing development effort, the
merchant experience will be purely through GQL/gRPC API and there will be 0 frontend for the merchant
experience, while frontend will be required for the customer experience.
● Frontend: Use **React (Typescript)** (do not use javascript, use typescript)
● Backend: Use **Golang** or **Rust**
● Protocol: Use **GraphQL** or **gRPC Web** (do not use REST or vanilla HTTP)
● Database: Use **PostgreSQL**
● Do NOT use an externally hosted database, this should be runnable from the local
environment.
● Cache: Use **Redis**
● Infrastructure: Use **Docker+Minikube** (local kubernetes) to run entire system locally
● Include README.md with clear and easy to understand instructions on how to run the system
locally
● Should not require us to find workarounds to get the system to run locally (we will fail the
submission if our graders need to look into your source codebase to find out how to run)
You are not allowed to use externally hosted services to implement any of the requirements in this document.
Do not submit your code on a public repo on github or any publicly accessible VCS. This coding challenge is
confidential.

## User Stories: Merchant

```
● As a merchant, via API call, I expect to be able to create pets to list on my store. A pet should have the
following information associated with it on creation:
● Name
● Species (Cat, Dog, or Frog)
● Age (number of human years)
● Picture of the pet
● Description
● Breeder Name (the name of the human who bred the pet)
● Breeder Email (the email of the human who bred the pet)
● Created At (the time the pet was created/listed on the store)
● As a merchant, via API call, I expect to be able to remove any pet from my store at any time before the
pet is purchased
● As a merchant, via API call, I can query which pets were sold (purchased) by an inclusive time range
(start and end date)
● In the API response I expect to see all details associated with each of the pet on creation
● As a merchant, via API call, I can query which pets have not been sold yet
● In the API response I expect to see all details associated with each of the pet on creation
```
## User Stories: Customer

```
● As a customer, I can open a dedicated url in my browser (you can choose a unique url on localhost),
and see a website for the merchant’s store. This website should show me the list of pets that are
available to purchase (should not show pets which are already purchased, browser page refresh is
sufficient to reflect this).
● As a customer, I can click a purchase button on any of the pets to instantly purchase the pet and take it
off the market (for free, we live in a world with no money for this exercise).
● I expect to see error messages if the pet is no longer available on the market to purchase. This
error message should be human readable.
● As a customer, I can click an add to cart button which adds the pets I want to buy to a cart. I can then
click a checkout button to instantly purchase all the pets in the cart.
```

```
● I expect to see error messages if the pets in my cart are no longer available on the market to
purchase. This error message should be human readable and include the names of the pets
which are no longer on the market (already purchased).
```
## Requirements Fullstack

```
● Performance: Frontend and backend should be performant (should have load times for all functionality
under 2 seconds for 1k concurrent users)
● Please implement a form of query pagination for the relevant endpoints which are impacted by
this performance metric.
● Implement a distributed cache (i.e. redis) to speed up reads on the backend
● Security:
● A merchant should only be able to access merchant endpoints for their store, and not merchant
endpoints for other stores (thus, each store will have its own separate collection of pets).
● All merchant and customer endpoints should be secured with basic HTTP authentication
standard (merchants should not be able to access customer endpoints and vice versa)
● There are other security requirements which are not explicitly called out here which should be
addressed for common web app attack vectors.
● Data should be encrypted at rest and in transit
● Sensitive information (you need to determine what is sensitive and what is not) should not be
stored in plain text
● Polish: User interface & user experience should be polished and professional, and the codebase should
be clean and easy to read.
● Race Conditions: For each functionality listed above in the user stories, please consider edge cases
around race conditions and put mechanisms in place to handle them gracefully.
```
## Submission Requirements

Anything else not specified in the requirements above are open to your choices and could be potential bonus
points. Please submit production ready code at the Sr. SWE level (however you may leave out unit test coverage
for sake of time).
● Include all source and configuration files for frontend, backend, and infrastructure
● Include a detailed README with instructions on how to run and use the system (see overview 5a)
● Include a video showing a working local system and usage of features described in the user stories above
● Every feature needs to be showcased on frontend and backend
● For backend, you can show DB values as they change and screen recordings of postman or curl
hitting your APIs for the merchant experience


