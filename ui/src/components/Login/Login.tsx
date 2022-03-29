import config from "../../config";
import { useState } from "react";

function Login() {
    const [password, setPassword] = useState("");

    const login_handler = async(e : any) => {
        e.preventDefault()
        try{
            let res = await fetch(`${config.endpoint}/api/public/session/new`, {
                method: "POST",
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json',
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
        <div className="flex justify-center items-center h-full w-full bg-gray-500">
            <div className="flex flex-col justify-center items-center w-26 p-14 bg-gray-700 shadow">
                <h2 className="pb-10 text-white">Admin area</h2>
                <form className="flex flex-col justify-center items-center" onSubmit={login_handler}>
                    <input 
                        className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" 
                        type="password" 
                        placeholder="Password" 
                        size={40}
                        onChange={(event) => { setPassword(event.target.value)}}/>
                    <br></br>
                    <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded w-50">Login</button>
                </form>
            </div>
        </div>
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