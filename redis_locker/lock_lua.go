package redis_locker

const (
	// 加锁
	lockScript = `
		local lock_key = KEYS[1]
		local lock_task = KEYS[2]
		local lock_value = ARGV[1]
		local lock_ttl = tonumber(ARGV[2])
		if redis.call('SET', lock_task, lock_value, 'NX', 'EX', lock_ttl + 3) then
			redis.call('SET', lock_key, '1', 'NX', 'EX', lock_ttl) 
			return "OK"
		end
		return nil
	`

	// 解锁
	unLockScript = `
		local lock_key = KEYS[1]
		local lock_value = ARGV[1]
		if redis.call('GET', lock_key) == lock_value then
			redis.call('DEL', lock_key)
			return "OK"
		else
			return nil
		end
	`

	// 续期
	renewScript = `
		local lock_key = KEYS[1]
		local lock_value = ARGV[1]
		local lock_ttl = tonumber(ARGV[2])
		local reentrant_key = lock_key .. ':count:' .. lock_value
		local reentrant_count = tonumber(redis.call('GET', reentrant_key) or '0')
		
		if reentrant_count > 0 or redis.call('GET', lock_key) == lock_value then
			redis.call('EXPIRE', lock_key, lock_ttl)
			redis.call('EXPIRE', reentrant_key, lock_ttl)
			return "OK"
		end
		
		return nil
	`
)
