# quotes

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
# add quote 
curl localhost:8080/quote/new -d '{"author": "test", "message": "test"}'
```