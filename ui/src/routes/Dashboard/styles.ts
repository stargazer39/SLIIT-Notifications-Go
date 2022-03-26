import styled from "styled-components";

export const Container = styled.div`
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    align-items: center;
`
interface DashItemProps {
    readonly green?: boolean;
    readonly red?: boolean;
    readonly pink?: boolean;
    readonly yellow?: boolean;
}

export const DashItem = styled.div<DashItemProps>`
    height: 225px;
    width: 225px;
    border-radius: 25px;
    padding: 10px;
    margin-bottom: 55px;
    margin-right: 55px;

    ${props => props.green && `
        border: #00fffc 2px solid;
        background-color: #007262;
        box-shadow: 0 0 13px 3px #006a53;
        color: white;`
    }
    
    ${props => props.red && `
        border: #ff0032 2px solid;
        background-color: #870505;
        box-shadow: 0 0 13px 3px #c50505;
        color: white;`
    }

    ${props => props.yellow && `
        border: #ffdd00 2px solid;
        background-color: #7c7302;
        box-shadow: 0 0 13px 3px #899d00;
        color: white;`
    }

    ${props => props.pink && `
        border: #ff00ee 2px solid;
        background-color: #720069;
        box-shadow: 0 0 13px 3px #52006a;
        color: white;`
    }

    transition: all 100ms;
    cursor: pointer;

    &:hover {
        transform: scale(1.1);
    }
`

export const FlexRow = styled.div`
    display: flex;
    flex-direction: row;
`

export const Title = styled.div`
    margin-top: 78px;
    margin-left: 104px;
    margin-bottom: 106px;
    font-size: 42px;
    font-weight: bold;
    color: white;
`

export const Wrapper = styled.div`
    margin: 0 20% 0 20%;
`