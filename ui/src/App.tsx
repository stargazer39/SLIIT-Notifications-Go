import { BrowserRouter, Outlet, Route, Routes } from "react-router-dom";
import Admin from "./routes/Admin/Admin";
import Home from "./routes/Home/Home";

function App(){
    return (
        <Outlet />
    )
}

export default App;