import { BrowserRouter, Outlet, Route, Routes } from "react-router-dom";
import Admin from "./components/Admin/Admin";
import Home from "./components/Home/Home";

function App(){
    return (
        <Outlet />
    )
}

export default App;