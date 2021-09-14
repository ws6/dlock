## dlock 
package is meant shared by all subpackages with distributed lock needs.

The current implementation of dlock is backing on redis.<br/>
Algo/protocol - https://redis.io/topics/distlock <br/>
Golang implementation - https://github.com/go-redsync/redsync <br/>
!!! We repurpose this library so we assume the lock is unlimited for current lock without explicitly calling Extend() <br/>

 
# configuration 
redis_config_string = the redisgo style configuration string <br/>
 addr,pool size,password,dbnum,IdleTimeout second <br/>
 e.g. <br/>
127.0.0.1:6379,100,password,0,30 <br/>
 or  <br/>
127.0.0.1:6379 without password <br/>
except addr, the other configuration is optional <br/>
default poolSize= 100 <br/>
default dbnum = 0 <br/>
default IdelTimeout Second = 0 <br/>
default password is empty <br/>

expire_second = the redis key expiry seconds.  defalt  15 <br/>
num_retry = number of retry when acquiring a lock. default 10 <br/>

# test setup
have environment variable <br/>
REDIS_CONFIG_STR = 127.0.0.1:6379,100,password,0,30 (see above redis_config_string) <br/>
