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