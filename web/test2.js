const axios = require('axios');
(async () => {
    try {
        const res = await axios.post('http://localhost:8080/v1/auth/login', {
            phone: '+77011234568',
            password: 'password123'
        });
        const token = res.data.access_token || (res.data.data && res.data.data.access_token);
        console.log("Logged in:", !!token);
        
        try {
            const res2 = await axios.get('http://localhost:8080/v1/profiles/me', { headers: { Authorization: "Bearer " + token } });
            console.log("Get Profile:", res2.status, res2.data);
        } catch (e2) {
             console.log("Get profile err:", e2.response ? e2.response.status + " " + JSON.stringify(e2.response.data) : e2.message);
        }
        
    } catch (e) {
        console.log("Login err:", e.response ? e.response.status + ' ' + JSON.stringify(e.response.data) : e.message);
    }
})();
