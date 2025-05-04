import React from 'react';
import {PanelRight} from "lucide-react";

interface HeaderProps {
    title: string;
}

export const Header: React.FC<HeaderProps> = ({ title }) => {
    return (
        <header className="bg-white border-b px-6 py-4">
            <div className="flex items-center justify-between">
                <PanelRight />
                <h1 className="text-2xl font-semibold text-gray-800">{title}</h1>
                <button className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
                    New Post
                </button>
            </div>
        </header>
    );
};