import autocannon from 'autocannon';

const instance = autocannon({
    url: 'http://localhost:8080/transactions',
    connections: 100,
    duration: 10,
    renderStatusCodes: true,

    requests: [
        {
            method: 'POST',
            path: '/transactions',
            headers: {
                'Content-Type': 'application/json'
            },
            setupRequest: (req) => {
                const userId = Math.floor(Math.random() * 1000) + 1;

                req.body = JSON.stringify({
                    user_id: userId,
                    name: 'loadtest',
                    amount: 1
                });

                return req;
            }
        }
    ]
});

autocannon.track(instance);




// random users hittig this transacation
// 1000 users with diff ids hit post
// concurently for 1000 users we add transactions redis check and postgres check 
