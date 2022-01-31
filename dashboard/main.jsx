const { useEffect } = React;
const { useState } = React;

function Home(props) {
    const [page, setPage] = useState(<UserManager />);

    const show_users = () => {
        setPage(<UserManager />)
    }

    const show_sites = () => {
        setPage(<SitesManager />)
    }

    return (
        <>
        <TopBar users={show_users} sites={show_sites}/>
        <div className="mt-5">{page}</div>
        </>
    );
}

function SitesManager(props) {
    const [sites, setSites] = useState([]);

    useEffect(() => {
        async function f(){
            try{
                const res = await fetch("/api/sites")
                let j = await res.json()
                setSites(j.sites)
            }catch(e){
                console.log(e)
            }
        }
        
        f()
    }, [])

    return (
        <div className="container">
            <h1>Site Manager</h1>
            <button type="button" className="btn btn-info my-2">Add New +</button>
            { 
                sites.map((s,i) => {
                    return <Site key={i} name={s.name} url={s.url} id={s.id} disabled={s.disabled}/>
                })
            }
        </div>
    )
}

function Site(props) {
    const [disabled, setDisabled] = useState(props.disabled)

    const disable_handler = async () => {
        try{
            let res = await fetch(`/api/sites/${props.id}/${disabled ? "enable" : "disable"}`)
            res = await res.json()

            if(res.error === true){
                return
            }

            setDisabled((prev) => {
                return !prev
            })
        }catch(e){
            console.log(e)
        }
    }
    return (
        <div className="site d-flex flex-row justify-content-center align-items-center mb-1 py-2 px-2">
            <div className="s-name px-2">{props.name}</div>
            <div className="spacer"></div>
            <div className="px-2">{props.url}</div>
            <button type="button" onClick={disable_handler} className={`btn ${disabled ? "btn-primary" : "btn-danger"}`}>{ !disabled ? "Disable" : "Enable"}</button>
        </div>
    )
}

function UserManager(props) {
    const [users, setUsers] = useState([]);
    const [addUser, setAddUser] = useState(false)

    useEffect(() => {
        async function f(){
            try{
                const res = await fetch("/api/users")
                let j = await res.json()
                setUsers(j.users)
            }catch(e){
                console.log(e)
            }
        }
        
        f()
    }, [])

    const add_user = () => {
        setAddUser((prev) => {
            return !prev
        })
    }

    return (
        <div className="container">
            <h1>User Manager</h1>
            <button type="button" disabled={addUser} className="btn btn-info my-2" onClick={add_user}>Add New +</button>
            {
                function (){
                    if(addUser){
                        return <AddUser />
                    }else{
                        " "
                    }
                }()
            }
            { 
                users.map((s,i) => {
                    return <User key={i} username={s.username} />
                })
            }
        </div>
    )
}

function User(props) {
    return (
        <div className="site d-flex flex-row justify-content-center align-items-center mb-1">
            <div className="s-name px-2">{props.username}</div>
            <div className="spacer"></div>
            <button type="button" className="btn btn-danger">Yeet</button>
        </div>
    )
}
function AddUser(props) {
    return (
        <div className="p-2 w-100 mb-1">
            <form>
                <div className="form-group">
                    <label for="username">Username</label>
                    <input type="text" className="form-control" id="username" placeholder="ITXXXXX"/>
                </div><br/>
                <div className="form-group">
                    <label for="pass">Password</label>
                    <input type="password" className="form-control" id="pass"/>
                </div><br/>
                <button type="button" className="btn btn-danger">Add</button>
                <button type="button" className="btn btn-danger mx-1">Cancel</button>
            </form>
        </div>
    )
}


function TopBar(props) {
    const [message, setMessage] = useState("")

    const showMessage = (msg) => {
        setMessage(msg)
        setTimeout(()=> {
            setMessage("")
        }, 10000)
    }

    const restart_bot = async () => {
        try{
            let res = await fetch("/api/bot/restart")
            res = await res.json()

            if(res.error === true){
                showMessage(`Error: ${res.message}`)
                return
            }

            showMessage("Success")
        }catch(e){
            showMessage(`Error: ${e}`)
        }
    }

    return (
        <div className="topbar px-4 d-flex flex-row align-items-center">
            <button onClick={props.users} className="btn btn-primary">Users</button>
            <button onClick={props.sites} className="btn btn-primary mx-2">Sites</button>
            <div className="spacer"></div>
            <div className="px-4">{message}</div>
            <button type="button" className="btn btn-danger" onClick={restart_bot}>Restart Bot</button>
        </div>
    )
}

ReactDOM.render(
    <Home />,
    document.getElementById('root'),
);