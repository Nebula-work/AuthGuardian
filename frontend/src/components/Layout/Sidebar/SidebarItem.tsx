import React, { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronDown, ChevronRight } from 'lucide-react';

interface SubItem {
    text: string;
    to: string;
}

interface SidebarItemProps {
    icon: React.ElementType;
    text: string;
    to: string;
    isCollapsed: boolean;
    children?: SubItem[];
}

export const SidebarItem: React.FC<SidebarItemProps> = ({
                                                            icon: Icon,
                                                            text,
                                                            to,
                                                            isCollapsed,
                                                            children = [],
                                                        }) => {
    const location = useLocation();
    const hasChildren = children.length > 0; // Check if there are child items
    const isParentActive = location.pathname === to;
    const isChildActive = children.some((child) => location.pathname === child.to);

    const [isOpen, setIsOpen] = useState(false);

    useEffect(() => {
        if (isChildActive) {
            setIsOpen(true); // Auto-expand if a child route is active
        }
    }, [isChildActive]);

    const handleClick = (e: React.MouseEvent) => {
        if (hasChildren) {
            e.preventDefault();
            setIsOpen((prev) => !prev); // Toggle the state for showing children
        }
    };

    const containerClasses = `flex items-center px-3 py-2 cursor-pointer rounded-lg ${
        isParentActive || isChildActive
            ? 'bg-indigo-50 text-indigo-600'
            : 'text-gray-600 hover:bg-gray-100'
    }`;

    return (
        <div>
            <Link to={hasChildren ? '#' : to} onClick={handleClick}>
                <div className={containerClasses}>
                    <Icon className="h-5 w-5 flex-shrink-0" />
                    {!isCollapsed && (
                        <>
                            <span className="ml-2 flex-1 truncate">{text}</span>
                            {hasChildren && (
                                <span className="ml-auto">
                                    {isOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                                </span>
                            )}
                        </>
                    )}
                </div>
            </Link>

            {!isCollapsed && hasChildren && isOpen && (
                <div className="ml-6 mt-1 space-y-1">
                    {children.map((child) => {
                        const isActive = location.pathname === child.to;
                        return (
                            <Link
                                key={child.to}
                                to={child.to}
                                className={`block px-3 py-2 text-sm rounded-lg ${
                                    isActive
                                        ? 'bg-indigo-50 text-indigo-600'
                                        : 'text-gray-600 hover:bg-gray-100'
                                }`}
                            >
                                {child.text}
                            </Link>
                        );
                    })}
                </div>
            )}
        </div>
    );
};
