import { Button, Container, LoginForm, TextInput, Title, Wrapper } from "./styles";
import config from "../../config";
import { useState } from "react";

function Login() {
    const [password, setPassword] = useState("");

    const login_handler = async(e : any) => {
        e.preventDefault()
        try{
            let res = await fetch(`${config.endpoint}/api/public/session/new`, {
                method: "POST",
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: "",
                    password
                })
            });
            let j : any = await res.json();

            console.log(j)

            if(j.error === false){
                
            }
        }catch(e){
            console.log(e)
        }
        return false
    }

    return (
        <Wrapper>
            <Container>
                    <Title>Admin area</Title>
                    <LoginForm onSubmit={login_handler}>
                        <TextInput type="password" placeholder="Password" onChange={(event) => { setPassword(event.target.value)}}/>
                        <Button>Login</Button>
                    </LoginForm>
            </Container>
        </Wrapper>
    )
}

function hash(s: string): Promise<string> {
    const utf8 = new TextEncoder().encode(s);
    return crypto.subtle.digest('SHA-256', utf8).then((hashBuffer) => {
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      const hashHex = hashArray
        .map((bytes) => bytes.toString(16).padStart(2, '0'))
        .join('');
      return hashHex;
    });
  }
export default Login;