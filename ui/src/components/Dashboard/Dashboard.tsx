import { Outlet, useNavigate } from "react-router-dom";
import { AiTwotoneHome } from 'react-icons/ai';

const Button = ({ children, onClick } : { children: any, onClick?: any }) => {
    return (
        <div 
            className="h-[5rem] w-[5rem] shadow mb-1 flex justify-center items-center bg-cyan-500 hover:bg-cyan-900 cursor-pointer"
            onClick={onClick}
        >
            {children}
        </div>
    )
}

function Dashboard() {
    const navigate = useNavigate();

    return (
        <div className="h-full w-full flex flex-row bg-gray-600">
            <div className="flex flex-col pl-1 pt-1">
                <Button onClick={() => { navigate("/admin/dashboard") }}>
                    <AiTwotoneHome className="object-contain w-full h-full p-4" color="white"/>
                </Button>
                <Button>
                    <AiTwotoneHome className="object-contain w-full h-full p-4" color="white" />
                </Button>
                <Button>
                    <AiTwotoneHome className="object-contain w-full h-full p-4" color="white" />
                </Button>
            </div>
            <div className="flex flex-col flex-grow-1">
                <div className="text-4xl py-4 pl-6 font-bold text-white shadow">DashBoard</div>
                <div className="pl-6 flex-grow-1 overflow-auto">
                    <Outlet />
                </div>
            </div>
        </div>
    )
}

export default Dashboard;