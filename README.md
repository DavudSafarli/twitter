## Twitter clone

The goal by building this project is to practice `Event-Sourcing`, `TDD`, and varios other technologies.
Techs that will be used in the project:
- Go
- Relational Database: Postgres
- Graph Database: Neo4j
- Search Engine: ElasticSearch
- Message Broker: Kafka
- Inmemory Store: Redis
  

I will do my best to make this a project that someone can look and learn something.

---

Functionalities that comes to my mind:
- Auth
  - Signup
  - Login
- Social network
  - Tweet
  - News Feed / Timeline
  - Like/Unlike a tweet
  - Send/Cancel a friend request
  - Accept/reject a friend request
  - Search users/tweets
---

For now, Phase 1 will include functionalities below:
## Phase 1
- Auth service
  - Signup
    - create a user with email, username and pwd
    - publish an event for `SearchIngestor`
    - publish an event for `SocialGraphBuilder`
  - Login
    - check username and password, return a token
      |   Path  |Method|
      |---------|------|
      | /signup | POST |
      | /login  | POST |
- SearchIngestor
  - listens to topics for various events like `UsedSignedUp`, `ProfileUpdated`, `TweetPosted` and feeds the search engine(ElasticSearch)
- SocialGraphBuilder
  - listens to topics for various events like `UsedSignedUp`, `Followed`, `Unfollowed` and builds the graph database(Neo4j)

