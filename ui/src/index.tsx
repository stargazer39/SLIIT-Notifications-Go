import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import reportWebVitals from './reportWebVitals';
import App from './App';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import Home from './components/Home/Home';
import Admin from './components/Admin/Admin';
import Dashboard from './components/Dashboard/Dashboard';
import Login from './components/Login/Login';
import DashHome from './components/DashHome/DashHome';
import SiteManager from './components/SiteManager/SiteManager';
import UserManager from './components/UserManager/UserManager';

import "bootstrap/dist/css/bootstrap.min.css";
import History from './components/History/History';

ReactDOM.render(
  <React.StrictMode>
    <BrowserRouter>
        <Routes>
          <Route path="/*" element={<App />}>
           <Route index element={<Home />} />
            <Route path="admin" element={<Admin />}>
              <Route path="dashboard" element={<Dashboard />}>
                <Route index element={<DashHome />}/>
                <Route path="sites" element={<SiteManager />}/>
                <Route path="users" element={<UserManager />}/>
              </Route>
             <Route path="login" element={<Login />} />
           </Route>
           <Route path="history/:id" element={<History />} />
           <Route path='*' element={<Navigate replace to="admin/login" />} />
          </Route>
        </Routes>
      </BrowserRouter>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
