import { Button, Container, LoginForm, TextInput, Title, Wrapper } from "./styles";

function Login() {
    return (
        <Wrapper>
            <Container>
                    <Title>Admin area</Title>
                    <LoginForm>
                        <TextInput type="password" placeholder="Password"/>
                        <Button>Login</Button>
                    </LoginForm>
            </Container>
        </Wrapper>
    )
}

export default Login;