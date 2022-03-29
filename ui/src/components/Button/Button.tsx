import styled from "styled-components";

export const Button = styled.button<{ red?:boolean, blue?:boolean }>`
    all: unset;
    min-height: 50px;
    min-width: 80px;
    text-align: center;
    border-radius: 5px;
    box-shadow: 0 0 5px 0 black;

    ${props => props.red && `
        background-color: red;
        color: white,
    `}

    ${props => props.blue && `
        background-color: blue;
        color: white;
    `}

    &:active {
        box-shadow: 0 0 0 0 black !important;
    }

    &:hover {
        backdrop-filter: contrast(40%);
    }
`