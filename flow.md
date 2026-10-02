main -> /transactions -> TransactionHandler 
handler.go ->GET  TransactionHandler -> storage.GetTransactions() 
storage.go -> storage struct which is pointer to sql db 
GetTransactions() -> s (storage instance) -> s.db.Select() -> wil execute query and insert value into db -> returns all transactions
    

handler.go -> POST TransactionHandler -> addTransactions  -> IMPORTANT 
this will validate the transaction and add it to the db 
decode json to go and insert value in t transactions
call-> transactionValidator -> check amount and name which is sent by user -> returns bool 

once we verify the transaction is valid we check userrules from redis cache 
its like say user pays via a qr to shop through hdfc before adding that transaction to hdfc db 
hdfc server will check the following things -> dailyLimit, spentToday, txLimit , err := getUserFromRules(t.UserID)


if err is redis.NIl -> which is redis does not have this value we need to pull it from postgres db 


redis cahes dailyLimit spentToday and txLimit NOT THE TRANSACTIONS 
so call goes to getUserRules (postgres to redis) return UserRules struct
this return rules not yet in redis in a variable 
-> s.db.Get( execute query with UserID as a param and put value into rules  )
Now spenttoday calcuates daily spent but since i needs all the transaction of we need to fetch it from postgres which leads to poor performance
QUERY -> 
SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		AND status = true
		AND created_at >= CURRENT_DATE

!!! NEED A FIX FOR THIS !!!

next call setUserRules ->  put rules into redis 
approved -> decideTransaction return bool

update value in redis using IncrByFloat 
CreateTransactions to add it in postgres

