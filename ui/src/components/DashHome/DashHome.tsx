import { Link, useNavigate } from "react-router-dom";

const DashItem = ({ children, ...rest } : { children: any, onClick?: any }) => {
    return (
        <div {...rest} className="flex m-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded h-[200px] w-[200px] justify-center items-center cursor-pointer">{children}</div>
    )
}

function DashHome() {
    const navigate = useNavigate();

    return (
        <div className="flex flex-col justify-center items-center">
            <div className="flex flex-row">
                <DashItem 
                    onClick={() => navigate("/admin/dashboard/sites")}>
                        Manage Sites
                </DashItem>
                <DashItem
                    onClick={() => navigate("/admin/dashboard/users")}>
                    Manage Users
                </DashItem>
            </div>
            <div className="flex flex-row">
                <DashItem>
                    Restart bot
                </DashItem>
                <DashItem>
                    Stats
                </DashItem>
            </div>
        </div>
    )
}

export default DashHome;