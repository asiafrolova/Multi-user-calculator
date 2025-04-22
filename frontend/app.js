let token = localStorage.getItem('token');
let tokenExpire = localStorage.getItem('tokenExpire');
let pollingInterval = null;

if (token && tokenExpire && Date.now() < parseInt(tokenExpire)) {
  document.getElementById('auth').style.display = 'none';
  document.getElementById('dashboard').style.display = 'block';
  document.getElementById('welcome').textContent = `Welcome back!`;
  startPolling();
} else {
  localStorage.removeItem('token');
  localStorage.removeItem('tokenExpire');
  token = null;
  tokenExpire = null;
}

function register() {
  const login = document.getElementById('register-login').value;
  const password = document.getElementById('register-password').value;

  fetch('http://localhost:8080/api/v1/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ login, password })
  }).then(res => res.json()).then(data => {
    alert('Registered successfully');
  });
}

function login() {
  const login = document.getElementById('login-login').value;
  const password = document.getElementById('login-password').value;

  fetch('http://localhost:8080/api/v1/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ login, password })
  }).then(res => res.json()).then(data => {
    token = data['access_token'];
    tokenExpire =  data.time * 1000;
    localStorage.setItem('token', token);
    localStorage.setItem('tokenExpire', tokenExpire);
    document.getElementById('auth').style.display = 'none';
    document.getElementById('dashboard').style.display = 'block';
    document.getElementById('welcome').textContent = `Welcome, ${login}`;
    startPolling();
  });
}

function submitExpression() {
  
  const expression = document.getElementById('expression').value;
  
  fetch('http://localhost:8080/api/v1/calculate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({ expression })
  }).then(() => {
    fetchExpressions();
  });
}

function fetchExpressions() {
  console.log(token)
  fetch('http://localhost:8080/api/v1/expressions', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  }).then(res => res.text()).then(text => {
    if (!text) return;
    const data = JSON.parse(text);
    const list = document.getElementById('expressions-list');
    list.innerHTML = '';
    data.expressions.forEach(expr => {
      const li = document.createElement('li');
      res = expr.result
      console.log(expr.status=="Completed",expr.status)
      if (expr.status!="Completed"){
        res="..."
      }
      li.textContent = `${expr.exp} = ${res} [${expr.status}]`;
      const delBtn = document.createElement('button');
      delBtn.textContent = 'Delete';
      delBtn.onclick = () => deleteExpression(expr.id);
      li.appendChild(delBtn);
      list.appendChild(li);
    });
  });
}

function startPolling() {
  fetchExpressions();
  if (pollingInterval) clearInterval(pollingInterval);
  pollingInterval = setInterval(() => {
    if (Date.now() >= parseInt(tokenExpire)) {
      logout();
    } else {
      fetchExpressions();
    }
  }, 5000);
}

function logout() {
  token = null;
  tokenExpire = null;
  localStorage.removeItem('token');
  localStorage.removeItem('tokenExpire');
  document.getElementById('auth').style.display = 'block';
  document.getElementById('dashboard').style.display = 'none';
  if (pollingInterval) clearInterval(pollingInterval);
  alert('Session expired or logged out. Please log in again.');
}

function deleteExpression(id) {
  fetch(`http://localhost:8080/api/v1/delete/expressions/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': 'Bearer ' + token
    }
  }).then(() => fetchExpressions());
}

function deleteAllExpressions() {
  fetch(`http://localhost:8080/api/v1/delete/expressions`, {
    method: 'DELETE',
    headers: {
      'Authorization': 'Bearer ' + token
    }
  }).then(() => fetchExpressions());
}

function deleteUser() {
  if (!confirm('Are you sure you want to delete your account?')) return;
  fetch('http://localhost:8080/api/v1/delete/user', {
    method: 'DELETE',
    headers: {
      'Authorization': 'Bearer ' + token
    }
  }).then(() => {
    logout();
    alert('Account deleted');
  });
}

// Добавим кнопку выхода
window.addEventListener('DOMContentLoaded', () => {
  const logoutBtn = document.createElement('button');
  logoutBtn.textContent = 'Выйти из аккаунта';
  logoutBtn.onclick = logout;
  document.getElementById('dashboard').appendChild(logoutBtn);
});