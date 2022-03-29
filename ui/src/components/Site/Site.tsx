import { useState } from "react";
import config from "../../config";
import { Button } from "../Button/Button";

type SiteProps = {
    name: string,
    url: string,
    id: string,
    disabled: boolean,
}

function Site(props: SiteProps) {
    const [disabled, setDisabled] = useState(props.disabled);

    const toggleState = () => {
        fetch(`${config.endpoint}/api/private/sites/${props.id}/${disabled ? "enable" : "disable"}`,{ 
            method: 'post',
            credentials: 'include'
        })
            .then(res => { console.log(res); return res.json() }) 
            .then(data => {
                if(data.error !== true){
                    setDisabled(!disabled);
                    return
                }
                console.log(data);
            })
            .catch(e => {
                console.error(e);
            })
    }

    return (
        <div className="flex flex-row mb-0.5 hover:bg-slate-600 bg-slate-900 text-white rounded py-2.5 px-3 shadow">
            <div>
                <div className="font-bold">{props.name}</div>
                <a className="text-sm font-light py-0.5 text-white" href={props.url}>
                    {props.url}
                </a>
            </div>
            
            <div className="flex-grow-1"></div>
            <button 
                onClick={toggleState}
                className={`${disabled ? `bg-red-500 hover:bg-red-600`:`bg-blue-500 hover:bg-blue-700`} text-white font-bold py-2 px-4 rounded`}
            >
                {disabled ? "Enable" : "Disable"}
            </button>
        </div>
    )
}

export default Site;