import styled from "styled-components";

export const Wrapper =styled.div`
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
`

export const Container = styled.div`
    width: 400px;
    dispaly: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding-bottom: 100px;
    border: 2px #00f9fd solid;
    border-radius: 12px;
    background: #4c4c4c;
    box-shadow: 0 0px 13px 11px #00394e;
`

export const TextInput = styled.input`
    padding: 10px 16px 10px 16px;
    font-size: 15px;
    width: 60%;
    border-radius: 25px;
    min-width: 350px;
`

export const LoginForm = styled.form`
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
`

export const Button = styled.button`
    padding: 10px 25px 10px 25px;
    margin-top: 15px;
    cursor: pointer;
    background-color: white;
    color: black;
    border: white 1px solid;
    border-radius: 40px;
`
export const Title = styled.div`
    padding: 86px 0 50px 0;
    font-size: 25px;
    text-align: center;
    color: white;
    font-weight: bold;
    font-size: 32px;
`