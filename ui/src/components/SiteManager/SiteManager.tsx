import Site from "../Site/Site";
import { useEffect, useState } from "react";
import config from "../../config";

function SiteManager() {
    const [sites, setSites] = useState<any[]>([]);

    useEffect(() => {
        fetch(`${config.endpoint}/api/private/sites`, {credentials: 'include'})
            .then(res => res.json())
            .then(data => {
                console.log(data);
                setSites(data.sites);
            })
            .catch(e => {
                console.error(e);
            }) 
    },[])

    return (
        <div className="flex flex-col">
            {
                sites.map((s) => { return <Site name={s.name} url={s.url} id={s.id} disabled={s.disabled}/> })
            }
        </div>
    )
}

export default SiteManager;