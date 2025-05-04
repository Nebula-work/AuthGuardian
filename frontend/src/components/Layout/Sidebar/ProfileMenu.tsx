import React, { useRef, useEffect } from 'react';
import { User, Settings, LogOut } from 'lucide-react';

interface ProfileMenuProps {
    isCollapsed: boolean;
    isProfileMenuOpen: boolean;
    setIsProfileMenuOpen: (value: boolean) => void;
}

export const ProfileMenu: React.FC<ProfileMenuProps> = ({
                                                            isCollapsed,
                                                            isProfileMenuOpen,
                                                            setIsProfileMenuOpen,
                                                        }) => {
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsProfileMenuOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [setIsProfileMenuOpen]);

    return (
        <div className="relative" ref={menuRef}>
            <div
                className="p-4  flex items-center cursor-pointer hover:bg-gray-50"
                onClick={() => setIsProfileMenuOpen(!isProfileMenuOpen)}
            >
                <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
                    <span className="text-sm font-medium text-gray-600">NA</span>
                </div>
                {!isCollapsed && (
                    <div className="ml-3 flex-1">
                        <p className="text-sm font-medium text-gray-700">Niladri Adak</p>
                        <p className="text-xs text-gray-500">niladri@example.com</p>
                    </div>
                )}
            </div>

            {/* Profile Menu with smooth transition */}
            <div
                className={`absolute bottom-full left-[100%] w-full bg-white border rounded-lg shadow-lg py-1 mb-2 ${
                    isProfileMenuOpen ? 'opacity-100 transform translate-y-0' : 'opacity-0 transform translate-y-2'
                } transition-all duration-200 ease-out`}
                style={{ pointerEvents: isProfileMenuOpen ? 'auto' : 'none' }}
            >
                <button className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center">
                    <User className="h-4 w-4 mr-2" />
                    Profile
                </button>
                <button className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50 flex items-center">
                    <Settings className="h-4 w-4 mr-2" />
                    Settings
                </button>
                <hr className="my-1" />
                <button className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-50 flex items-center">
                    <LogOut className="h-4 w-4 mr-2" />
                    Logout
                </button>
            </div>
        </div>
    );
};
