# go-a/b

A simple a/b metrics data collection and visualization with redis.

## Example loading lua script from golang

```golang
func () {
  //
	script1 := client.ScriptLoad(`local value = cmsgpack.pack({ARGV[2], ARGV[1]})
redis.call('zadd', KEYS[1], ARGV[1], value)
return redis.status_reply('ok')`)

	if script1.Err() != nil {
		log.Println("error loading script:", script1.Err().Error())
	}
	// Else, get the sha
	sha := script1.Val()
	log.Println(sha)

	cmd2 := client.EvalSha(sha, []string{"test:1"}, "1421481600000", "yes")
}
```


## Bandit algorithm requirements


As a user, <br>
I want to create a new bandit test, <br>
In order to test the performance of the ui.

|------------------------------|
| Feature      | Pull | Reward |
|--------------|------|--------|
| Red button   | 0    | 0      |
| Green button | 0    | 0      |
| Blue button  | 0    | 0      |
|------------------------------|

As a user, <br>
I want to set the initial values of the bandit, <br>
In order to have better control over the exploitation.

```bash
$ PUT /bandits/experiment-name

{
	"values": [0,1,2,3]
}
```

- upload data 
- download data
- replay from last
- reset from start
- duplicate
- getURL
- update experiment name
- update labels
- set labels
- get reward/pull ratio
- store in memory/redis/dedicated storage?


Frontend guide

1. Call the api to select an arm
2. The call should return the arm, the feature information and some metadata
3. The data should be cached locally so that the user does not need to call the cached data again (persistent)
4. If the user clicks it, then update it with the score 1
5. Else if the user navigates out of the page without doing anything, then update it with the score 0