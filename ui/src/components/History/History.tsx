import React from "react";
import { useParams } from "react-router-dom";

function History() {
    const params = useParams();

    return (
        <div className="h-full w-full">
            <div className="text-4xl py-4 pl-6 font-bold text-black shadow">History</div>
            <div className="p-2">
                {params.id}
            </div>
        </div>
    )
}

export default History;