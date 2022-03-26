import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import reportWebVitals from './reportWebVitals';
import App from './App';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import Home from './routes/Home/Home';
import Admin from './routes/Admin/Admin';
import Dashboard from './routes/Dashboard/Dashboard';
import Login from './routes/Login/Login';
import DashHome from './routes/Dashboard/DashHome';
import SiteManager from './routes/Dashboard/SiteManager';

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
            </Route>
             <Route path="login" element={<Login />} />
           </Route>
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
