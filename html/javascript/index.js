function myFunction() {
    var username = String(document.getElementById("CreateUser-Username").value);
    var password = String(document.getElementById("CreateUser-Password").value);
    var email = String(document.getElementById("CreateUser-Email").value);
    alert(username)
    fetch('http://localhost:8080/api/users', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({ 
        'email': email,
        'password': password,
        'username': username
       })
    })
  }