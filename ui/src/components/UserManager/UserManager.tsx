import Site from "../Site/Site";
import { useEffect, useState } from "react";
import config from "../../config";
import User from "../User/User";

function UserManager() {
    const [users, setUsers] = useState<any[]>([]);

    useEffect(() => {
        fetch(`${config.endpoint}/api/private/users`, {credentials: 'include'})
            .then(res => res.json())
            .then(data => {
                console.log(data);
                setUsers(data.users);
            })
            .catch(e => {
                console.error(e);
            }) 
    },[])

    return (
        <div className="flex flex-col">
            {
                users.map((s) => { return <User username={s.username} id={s.id} nick={"Nick"} disabled={s.disabled}/> })
            }
        </div>
    )
}

export default UserManager;