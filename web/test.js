const go = async () => {
    const num = '+770' + Math.floor(10000000 + Math.random() * 90000000);
    const res1 = await fetch('http://localhost:8080/v1/auth/register', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({phone: num, password: 'password123'})
    });
    const data1 = await res1.json();
    let token = data1.data.access_token;
    
    // exact payload that frontend sends based on cleanData
    const payload = {
      display_name: 'Alex',
      bio: '',
      city: '',
    };
    const res2 = await fetch('http://localhost:8080/v1/profiles/me', {
        method: 'PUT',
        headers: {'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token},
        body: JSON.stringify(payload)
    });
    console.log('Profile PUT Status:', res2.status, await res2.text());
}
go();
