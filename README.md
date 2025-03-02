# quotes

## Ideas
- New quote submissions are approved by the community. Some percentage of the community must validate/approve before a submissions is accepted and published. 
## API Interactions
```bash
# add category
curl localhost:8080/category/new -d '{"name": "motivational"}'
# add quote 
curl localhost:8080/quote/new -d '{"author": "test", "message": "test"}'
```