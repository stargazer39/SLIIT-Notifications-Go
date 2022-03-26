import { Outlet } from "react-router-dom";
import { Wrapper } from "./styles";

function Dashboard() {
    return (
        <> 
            <Wrapper>
                <Outlet />
            </Wrapper>
        </>
    )
}

export default Dashboard;