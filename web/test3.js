const axios = require('axios');
(async () => {
    try {
        const res = await axios.post('http://localhost:8080/v1/auth/login', {
            phone: '+77011234568',
            password: 'password123'
        });
        const token = res.data.access_token || (res.data.data && res.data.data.access_token);
        
        try {
            const putRes = await axios.put('http://localhost:8080/v1/profiles/me', {
                display_name: 'Test Profile',
                gender: 'male',
                looking_for: 'female'
            }, { headers: { Authorization: "Bearer " + token } });
            console.log("Put Profile:", putRes.status);
            
            const getRes = await axios.get('http://localhost:8080/v1/profiles/me', { headers: { Authorization: "Bearer " + token } });
            console.log("Get Profile:", getRes.status, !!getRes.data);
            
        } catch (e2) {
             console.log("Profile action err:", e2.response ? e2.response.status + " " + JSON.stringify(e2.response.data) : e2.message);
        }
    } catch (e) { }
})();
