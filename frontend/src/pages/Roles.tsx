import React, { useState, useEffect } from 'react';
import api from '../services/api';
import { Role, Permission, Organization } from '../types';

const Roles: React.FC = () => {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit] = useState(10);
  const [total, setTotal] = useState(0);
  const [selectedOrgId, setSelectedOrgId] = useState('');

  // Form state for creating/editing roles
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    description: '',
    organizationId: '',
    permissionIds: [] as string[],
  });

  // Permission management modal
  const [isPermissionModalOpen, setIsPermissionModalOpen] = useState(false);
  const [selectedRoleId, setSelectedRoleId] = useState('');
  const [rolePermissions, setRolePermissions] = useState<Permission[]>([]);
  const [availablePermissions, setAvailablePermissions] = useState<Permission[]>([]);
  const [selectedPermissionIds, setSelectedPermissionIds] = useState<string[]>([]);

  useEffect(() => {
    fetchData();
  }, [currentPage, selectedOrgId]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const skip = (currentPage - 1) * limit;
      let url = `/api/roles?limit=${limit}&skip=${skip}`;

      if (selectedOrgId) {
        url += `&organizationId=${selectedOrgId}`;
      }

      const [rolesResponse, orgsResponse] = await Promise.all([
        api.get(url),
        api.get('/api/organizations')
      ]);

      if (rolesResponse.data.success) {
        setRoles(rolesResponse.data.data.roles);
        setTotal(rolesResponse.data.data.total);
      } else {
        setError('Failed to fetch roles');
      }

      if (orgsResponse.data.success) {
        setOrganizations(orgsResponse.data.data.organizations);
      }

      try {
        // Fetch permissions from the API
        const permissionsResponse = await api.get('/api/permissions');
        if (permissionsResponse.data.success) {
          setPermissions(permissionsResponse.data.data.permissions);
        } else {
          console.error('Failed to fetch permissions from API:', permissionsResponse.data.error);
          setError('Failed to load permissions. Please try again.');
        }
      } catch (permErr) {
        console.error('Error fetching permissions:', permErr);
        setError('Failed to load permissions. Please try again.');
      }
    } catch (err) {
      setError('Error loading data. Please try again.');
      console.error('Error fetching data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this role?')) return;

    try {
      const response = await api.delete(`/api/roles/${id}`);
      if (response.data.success) {
        setRoles(roles.filter(role => role.id !== id));
        alert('Role deleted successfully');
      } else {
        setError('Failed to delete role');
      }
    } catch (err) {
      setError('Error deleting role. Please try again.');
      console.error('Error deleting role:', err);
    }
  };

  const openCreateModal = () => {
    setIsEditing(false);
    setFormData({
      id: '',
      name: '',
      description: '',
      organizationId: '',
      permissionIds: []
    });
    setIsModalOpen(true);
  };

  const openEditModal = (role: Role) => {
    setIsEditing(true);
    setFormData({
      id: role.id,
      name: role.name,
      description: role.description,
      organizationId: role.organizationId || '',
      permissionIds: role.permissionIds || []
    });
    setIsModalOpen(true);
  };

  const openPermissionModal = async (role: Role) => {
    setSelectedRoleId(role.id);

    // Find permissions for this role
    const rolePermIds = role.permissionIds || [];
    const rolePerm = permissions.filter(p => rolePermIds.includes(p.id));
    setRolePermissions(rolePerm);

    // Get available permissions that can be added
    const availablePerm = permissions.filter(p => !rolePermIds.includes(p.id));
    setAvailablePermissions(availablePerm);

    setSelectedPermissionIds([]);
    setIsPermissionModalOpen(true);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSelectMultiple = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const { name, options } = e.target;
    const values: string[] = [];

    for (let i = 0; i < options.length; i++) {
      if (options[i].selected) {
        values.push(options[i].value);
      }
    }

    setFormData({ ...formData, [name]: values });
  };

  const handlePermissionSelection = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const options = e.target.options;
    const values: string[] = [];

    for (let i = 0; i < options.length; i++) {
      if (options[i].selected) {
        values.push(options[i].value);
      }
    }

    setSelectedPermissionIds(values);
  };

  const handleAddPermissions = async () => {
    if (selectedPermissionIds.length === 0) return;

    try {
      const response = await api.post(`/api/roles/${selectedRoleId}/permissions`, {
        permissionIds: selectedPermissionIds
      });

      if (response.data.success) {
        // Update local state
        const updatedRole = response.data.data;
        setRoles(roles.map(r => r.id === selectedRoleId ? updatedRole : r));

        // Update permissions lists in the modal
        const updatedRolePermIds = updatedRole.permissionIds || [];
        const rolePerm = permissions.filter(p => updatedRolePermIds.includes(p.id));
        setRolePermissions(rolePerm);

        const availablePerm = permissions.filter(p => !updatedRolePermIds.includes(p.id));
        setAvailablePermissions(availablePerm);

        setSelectedPermissionIds([]);
        alert('Permissions added successfully');
      } else {
        setError(response.data.error || 'Failed to add permissions');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred. Please try again.');
      console.error('Error adding permissions:', err);
    }
  };

  const handleRemovePermission = async (permissionId: string) => {
    try {
      const response = await api.delete(`/api/roles/${selectedRoleId}/permissions`, {
        data: { permissionIds: [permissionId] }
      });

      if (response.data.success) {
        // Update local state
        const updatedRole = response.data.data;
        setRoles(roles.map(r => r.id === selectedRoleId ? updatedRole : r));

        // Update permissions lists in the modal
        const updatedRolePermIds = updatedRole.permissionIds || [];
        const rolePerm = permissions.filter(p => updatedRolePermIds.includes(p.id));
        setRolePermissions(rolePerm);

        const availablePerm = permissions.filter(p => !updatedRolePermIds.includes(p.id));
        setAvailablePermissions(availablePerm);

        alert('Permission removed successfully');
      } else {
        setError(response.data.error || 'Failed to remove permission');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred. Please try again.');
      console.error('Error removing permission:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      let response;

      if (isEditing) {
        response = await api.put(`/api/roles/${formData.id}`, formData);
      } else {
        response = await api.post('/api/roles', formData);
      }

      if (response.data.success) {
        setIsModalOpen(false);
        fetchData(); // Refresh the roles list
        alert(isEditing ? 'Role updated successfully' : 'Role created successfully');
      } else {
        setError(response.data.error || 'Operation failed');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'An error occurred. Please try again.');
      console.error('Error saving role:', err);
    }
  };

  const totalPages = Math.ceil(total / limit);

  return (
      <div className="container mx-auto px-4 py-6">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-semibold text-gray-900">Roles</h1>
          <button
              onClick={openCreateModal}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Add New Role
          </button>
        </div>

        {error && (
            <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path
                        fillRule="evenodd"
                        d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                        clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              </div>
            </div>
        )}

        <div className="bg-white p-4 rounded-lg shadow mb-6">
          <div className="flex items-center space-x-4">
            <label htmlFor="organizationFilter" className="block text-sm font-medium text-gray-700">
              Filter by Organization:
            </label>
            <select
                id="organizationFilter"
                value={selectedOrgId}
                onChange={(e) => setSelectedOrgId(e.target.value)}
                className="block w-64 border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
            >
              <option value="">All Organizations</option>
              {organizations.map(org => (
                  <option key={org.id} value={org.id}>{org.name}</option>
              ))}
            </select>
          </div>
        </div>

        {loading ? (
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
            </div>
        ) : (
            <>
              <div className="overflow-x-auto bg-white shadow-md rounded-lg">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Organization</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">System Default</th>
                    <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                  {roles.length > 0 ? (
                      roles.map((role) => (
                          <tr key={role.id}>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <div className="text-sm font-medium text-gray-900">{role.name}</div>
                            </td>
                            <td className="px-6 py-4">
                              <div className="text-sm text-gray-900">{role.description}</div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <div className="text-sm text-gray-900">
                                {organizations.find(o => o.id === role.organizationId)?.name || 'System-wide'}
                              </div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${role.isSystemDefault ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800'}`}>
                          {role.isSystemDefault ? 'Yes' : 'No'}
                        </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                              <button
                                  onClick={() => openEditModal(role)}
                                  className="text-indigo-600 hover:text-indigo-900 mr-3"
                                  disabled={role.isSystemDefault}
                              >
                                Edit
                              </button>
                              <button
                                  onClick={() => openPermissionModal(role)}
                                  className="text-green-600 hover:text-green-900 mr-3"
                              >
                                Permissions
                              </button>
                              <button
                                  onClick={() => handleDelete(role.id)}
                                  className="text-red-600 hover:text-red-900"
                                  disabled={role.isSystemDefault}
                              >
                                Delete
                              </button>
                            </td>
                          </tr>
                      ))
                  ) : (
                      <tr>
                        <td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500">
                          No roles found
                        </td>
                      </tr>
                  )}
                  </tbody>
                </table>
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                  <div className="flex items-center justify-between px-4 py-3 bg-white border-t border-gray-200 sm:px-6 mt-4 rounded-lg">
                    <div className="flex-1 flex justify-between sm:hidden">
                      <button
                          onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                          disabled={currentPage === 1}
                          className={`relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md ${
                              currentPage === 1 ? 'bg-gray-100 text-gray-400' : 'bg-white text-gray-700 hover:bg-gray-50'
                          }`}
                      >
                        Previous
                      </button>
                      <button
                          onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                          disabled={currentPage === totalPages}
                          className={`ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md ${
                              currentPage === totalPages ? 'bg-gray-100 text-gray-400' : 'bg-white text-gray-700 hover:bg-gray-50'
                          }`}
                      >
                        Next
                      </button>
                    </div>
                    <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                      <div>
                        <p className="text-sm text-gray-700">
                          Showing <span className="font-medium">{(currentPage - 1) * limit + 1}</span> to{' '}
                          <span className="font-medium">{Math.min(currentPage * limit, total)}</span> of{' '}
                          <span className="font-medium">{total}</span> results
                        </p>
                      </div>
                      <div>
                        <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                          <button
                              onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                              disabled={currentPage === 1}
                              className={`relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium ${
                                  currentPage === 1 ? 'text-gray-300' : 'text-gray-500 hover:bg-gray-50'
                              }`}
                          >
                            <span className="sr-only">Previous</span>
                            <svg className="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                              <path fillRule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clipRule="evenodd" />
                            </svg>
                          </button>

                          {Array.from({ length: Math.min(5, totalPages) }).map((_, i) => {
                            let pageNum = i + 1;
                            if (totalPages > 5 && currentPage > 3) {
                              pageNum = Math.min(currentPage - 3 + i + 1, totalPages);
                            }

                            return (
                                <button
                                    key={pageNum}
                                    onClick={() => setCurrentPage(pageNum)}
                                    className={`relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium ${
                                        currentPage === pageNum ? 'bg-blue-50 border-blue-500 text-blue-600' : 'bg-white text-gray-700 hover:bg-gray-50'
                                    }`}
                                >
                                  {pageNum}
                                </button>
                            );
                          })}

                          <button
                              onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                              disabled={currentPage === totalPages}
                              className={`relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium ${
                                  currentPage === totalPages ? 'text-gray-300' : 'text-gray-500 hover:bg-gray-50'
                              }`}
                          >
                            <span className="sr-only">Next</span>
                            <svg className="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                              <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                            </svg>
                          </button>
                        </nav>
                      </div>
                    </div>
                  </div>
              )}
            </>
        )}

        {/* Role Create/Edit Modal */}
        {isModalOpen && (
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-10">
              <div className="bg-white rounded-lg max-w-md w-full mx-4 p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold">{isEditing ? 'Edit Role' : 'Create Role'}</h2>
                  <button
                      onClick={() => setIsModalOpen(false)}
                      className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                <form onSubmit={handleSubmit}>
                  <div className="space-y-4">
                    <div>
                      <label htmlFor="name" className="block text-sm font-medium text-gray-700">Name</label>
                      <input
                          type="text"
                          id="name"
                          name="name"
                          value={formData.name}
                          onChange={handleInputChange}
                          required
                          className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      />
                    </div>

                    <div>
                      <label htmlFor="description" className="block text-sm font-medium text-gray-700">Description</label>
                      <textarea
                          id="description"
                          name="description"
                          value={formData.description}
                          onChange={handleInputChange}
                          rows={3}
                          className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      />
                    </div>

                    <div>
                      <label htmlFor="organizationId" className="block text-sm font-medium text-gray-700">Organization</label>
                      <select
                          id="organizationId"
                          name="organizationId"
                          value={formData.organizationId}
                          onChange={handleInputChange}
                          className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      >
                        <option value="">System-wide</option>
                        {organizations.map(org => (
                            <option key={org.id} value={org.id}>{org.name}</option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label htmlFor="permissionIds" className="block text-sm font-medium text-gray-700">Permissions</label>
                      <select
                          id="permissionIds"
                          name="permissionIds"
                          multiple
                          value={formData.permissionIds}
                          onChange={handleSelectMultiple}
                          className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                          style={{ minHeight: '120px' }}
                      >
                        {permissions.map(permission => (
                            <option key={permission.id} value={permission.id}>
                              {permission.name} - {permission.description}
                            </option>
                        ))}
                      </select>
                      <p className="mt-1 text-xs text-gray-500">Hold Ctrl/Cmd to select multiple</p>
                    </div>
                  </div>

                  <div className="mt-6 flex justify-end">
                    <button
                        type="button"
                        onClick={() => setIsModalOpen(false)}
                        className="mr-3 px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                    >
                      Cancel
                    </button>
                    <button
                        type="submit"
                        className="px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                    >
                      {isEditing ? 'Update' : 'Create'}
                    </button>
                  </div>
                </form>
              </div>
            </div>
        )}

        {/* Permission Management Modal */}
        {isPermissionModalOpen && (
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-10">
              <div className="bg-white rounded-lg max-w-3xl w-full mx-4 p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold">Manage Role Permissions</h2>
                  <button
                      onClick={() => setIsPermissionModalOpen(false)}
                      className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Add Permissions</h3>
                  <div className="flex items-start space-x-3">
                    <select
                        multiple
                        value={selectedPermissionIds}
                        onChange={handlePermissionSelection}
                        className="block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                        style={{ minHeight: '150px' }}
                    >
                      {availablePermissions.map(permission => (
                          <option key={permission.id} value={permission.id}>
                            {permission.name} - {permission.description}
                          </option>
                      ))}
                    </select>
                    <button
                        type="button"
                        onClick={handleAddPermissions}
                        disabled={selectedPermissionIds.length === 0}
                        className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Add &rarr;
                    </button>
                  </div>
                  <p className="mt-1 text-xs text-gray-500">Hold Ctrl/Cmd to select multiple</p>
                </div>

                <h3 className="text-lg font-medium text-gray-900 mb-2">Current Permissions</h3>
                {rolePermissions.length > 0 ? (
                    <div className="overflow-x-auto bg-white border border-gray-200 rounded-lg">
                      <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                        <tr>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Resource</th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                          <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                        {rolePermissions.map((permission) => (
                            <tr key={permission.id}>
                              <td className="px-6 py-4 whitespace-nowrap">
                                <div className="text-sm font-medium text-gray-900">{permission.name}</div>
                              </td>
                              <td className="px-6 py-4">
                                <div className="text-sm text-gray-900">{permission.description}</div>
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap">
                                <div className="text-sm text-gray-900">{permission.resource}</div>
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap">
                                <div className="text-sm text-gray-900">{permission.action}</div>
                              </td>
                              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                <button
                                    onClick={() => handleRemovePermission(permission.id)}
                                    className="text-red-600 hover:text-red-900"
                                >
                                  Remove
                                </button>
                              </td>
                            </tr>
                        ))}
                        </tbody>
                      </table>
                    </div>
                ) : (
                    <p className="text-gray-500">No permissions assigned to this role yet.</p>
                )}

                <div className="mt-6 flex justify-end">
                  <button
                      type="button"
                      onClick={() => setIsPermissionModalOpen(false)}
                      className="px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
        )}
      </div>
  );
};

export default Roles;
