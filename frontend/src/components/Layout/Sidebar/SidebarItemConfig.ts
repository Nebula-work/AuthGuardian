import {
    LayoutDashboard,
    Users,
    Building2,
    ShieldCheck,
    FileText
} from 'lucide-react';

export const sidebarItems = [
    {
        icon: LayoutDashboard,
        text: 'Dashboard',
        to: '/',
        children: [],
    },
    {
        icon: Users,
        text: 'Users',
        to: '/users',
        children: [],
    },
    {
        icon: Building2,
        text: 'Organizations',
        to: '/organizations',
        children: [],
    },
    {
        icon: ShieldCheck,
        text: 'Roles',
        to: '/roles',
        children: [],
    },
    {
        icon: FileText,
        text: 'Documentation',
        to: '/docs',
        children: [],
    }
];