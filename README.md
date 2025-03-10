# quotes

## Store
Using postgres running in a docker container at the moment. 
## Ideas
- New quote submissions are approved by the community. Some percentage of the community must validate/approve before a submissions is accepted and published. 
## API Interactions
```bash
# healthcheck
➜  quotes-site git:(main) ✗ curl localhost:8080/health | jq .
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    28  100    28    0     0  19047      0 --:--:-- --:--:-- --:--:-- 28000
{
  "msg": "system is healthy"
}
# add category
curl localhost:8080/category/new -d '{"name": "motivational"}'
# get all categories 
curl localhost:8080/category | jq .
# add author
curl localhost:8080/author/new -d '{"name": "einstein"}'
# get authors
curl localhost:8080/author
# add quote 
curl localhost:8080/quote/new -d '{"category": {"id": 1},"author": {"id": 1}, "message": "There are only two ways to live your life. One is as though nothing is a miracle. The other is as though everything is a miracle."}'
“There are only two ways to live your life. One is as though nothing is a miracle. The other is as though everything is a miracle.”
"Now here you see it takes all the running you can do to stay in the same place. If you want to get somewhere else, , you must run at least twice as fast as that!”
― Lewis Carroll, Alice Through The Looking Glass"
```

## Queries

```sql
SELECT
  c.name,
  a.name,
  m.message
FROM messages m
JOIN categories c ON c.id = m.category_id
JOIN authors a ON a.id = m.author_id;
```
