import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import { Role, Permission } from '../types';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch roles
        const rolesResponse = await api.get('/api/roles');
        setRoles(rolesResponse.data.data || []);
        
        // Fetch permissions
        const permissionsResponse = await api.get('/api/roles/1/permissions');
        setPermissions(permissionsResponse.data.data || []);
        
        setLoading(false);
      } catch (err) {
        console.error('Error fetching dashboard data:', err);
        setError('Failed to load dashboard data');
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-blue-500">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded my-4">
        {error}
      </div>
    );
  }

  // Create a Set of user's role IDs for faster lookups
  const userRoleIds = new Set(user?.roleIds || []);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Dashboard</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Your Information</h2>
          {user ? (
            <div>
              <p className="mb-2"><span className="font-medium">Username:</span> {user.username}</p>
              <p className="mb-2"><span className="font-medium">Email:</span> {user.email}</p>
              <p className="mb-2"><span className="font-medium">Name:</span> {user.firstName} {user.lastName}</p>
              <p className="mb-2"><span className="font-medium">Account Status:</span> {user.active ? 'Active' : 'Inactive'}</p>
              <p className="mb-2"><span className="font-medium">Email Verified:</span> {user.emailVerified ? 'Yes' : 'No'}</p>
            </div>
          ) : (
            <p>User information not available</p>
          )}
        </div>
        
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Your Roles</h2>
          {roles.length > 0 ? (
            <ul className="divide-y divide-gray-200">
              {roles.filter(role => userRoleIds.has(role.id)).map((role) => (
                <li key={role.id} className="py-2">
                  <div className="font-medium">{role.name}</div>
                  <div className="text-sm text-gray-600">{role.description}</div>
                </li>
              ))}
            </ul>
          ) : (
            <p>No roles assigned</p>
          )}
        </div>
      </div>
      
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold mb-4">System Statistics</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg">
            <div className="text-3xl font-bold text-blue-600">{roles.length}</div>
            <div className="text-sm text-blue-900">Total Roles</div>
          </div>
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="text-3xl font-bold text-green-600">{permissions.length}</div>
            <div className="text-sm text-green-900">Total Permissions</div>
          </div>
          <div className="bg-purple-50 p-4 rounded-lg">
            <div className="text-3xl font-bold text-purple-600">{user?.organizationIds?.length || 0}</div>
            <div className="text-sm text-purple-900">Your Organizations</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;