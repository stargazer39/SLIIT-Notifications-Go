import { Outlet } from "react-router-dom";
import { Container } from "./styles";

function Admin(){
    return (
        <Container>
            <Outlet />
        </Container>
    )
}

export default Admin;