import { Link } from "react-router-dom";
import { Container, DashItem, FlexRow, Title, Wrapper } from "./styles";

function DashHome() {
    return (
        <> 
            <Title>Dashboard</Title>
            <Container>
                <FlexRow>
                    <DashItem green>Manage Sites</DashItem>
                    <DashItem red>Manage Users</DashItem>
                </FlexRow>
                <FlexRow>
                    <DashItem pink>Restart Bot</DashItem>
                    <DashItem yellow>Stats</DashItem>
                </FlexRow>
            </Container>
        </>
    )
}

export default DashHome;