import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { PlusIcon, UserGroupIcon, UserPlusIcon } from '@heroicons/react/24/outline';
import { toast } from 'react-hot-toast';
import { apiService } from '../../services/api';
import type { Team } from '../../types';
import Table from '../UI/Table';
import Button from '../UI/Button';
import Modal from '../UI/Modal';
import { formatDate, exportToCSV } from '../../utils';
import { useAppStore } from '../../store';

const Teams: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { language, selectedRows, setSelectedRows, clearSelectedRows } = useAppStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
  const [isDeactivateModalOpen, setIsDeactivateModalOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [newTeamName, setNewTeamName] = useState('');
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [usersToDeactivate, setUsersToDeactivate] = useState<Set<number>>(new Set());

  // Fetch teams
  const { data: teams, isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: () => apiService.getTeams(),
  });

  // Fetch users for adding to team
  const { data: users } = useQuery({
    queryKey: ['users', 'active'],
    queryFn: () => apiService.getUsers({ isActive: true }),
  });

  // Fetch users for selected team when deactivation modal opens
  const { data: teamUsers } = useQuery({
    queryKey: ['users', 'team', selectedTeam?.id],
    queryFn: () => apiService.getUsers({ teamId: selectedTeam?.id, isActive: true }),
    enabled: !!selectedTeam?.id && isDeactivateModalOpen,
  });

  // Ensure arrays are never null/undefined
  const teamsData = teams || [];
  const usersData = users || [];
  const teamUsersData = teamUsers || [];

  // Create team mutation
  const createTeamMutation = useMutation({
    mutationFn: (name: string) => apiService.createTeam({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      toast.success(t('messages.teamCreated'));
      setIsCreateModalOpen(false);
      setNewTeamName('');
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Add user to team mutation
  const addUserToTeamMutation = useMutation({
    mutationFn: ({ teamId, userId }: { teamId: number; userId: number }) =>
      apiService.addUserToTeam(teamId, { userId: userId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User added to team');
      setIsAddUserModalOpen(false);
      setSelectedUserId(null);
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Deactivate users mutation
  const deactivateUsersMutation = useMutation({
    mutationFn: ({ teamId, userIds }: { teamId: number; userIds: number[] }) =>
      apiService.deactivateUsers(teamId, { userIds: userIds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success(t('messages.usersDeactivated'));
      setIsDeactivateModalOpen(false);
      setUsersToDeactivate(new Set());
      clearSelectedRows();
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  // Delete teams mutation
  const deleteTeamsMutation = useMutation({
    mutationFn: async (teamIds: number[]) => {
      await Promise.all(teamIds.map(id => apiService.deleteTeam(id)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success(t('messages.teamsDeleted') || 'Teams deleted successfully');
      clearSelectedRows();
    },
    onError: () => {
      toast.error(t('messages.errorOccurred'));
    },
  });

  const handleExport = () => {
    const data = teamsData.map(team => ({
      id: team.id,
      name: team.name,
      createdAt: formatDate(team.createdAt, language),
      updatedAt: formatDate(team.updatedAt, language),
    }));
    exportToCSV(data, ['id', 'name', 'createdAt', 'updatedAt'], 'teams.csv');
    toast.success(t('messages.exportSuccess'));
  };

  const columns = [
    {
      key: 'id',
      header: 'ID',
      accessor: (team: Team) => team.id,
      searchValue: (team: Team) => `${team.id} ${team.name}`,
      sortValue: (team: Team) => team.id,
      sortable: true,
      width: 'w-16',
    },
    {
      key: 'name',
      header: t('teams.teamName'),
      accessor: (team: Team) => (
        <div className="flex items-center gap-2">
          <UserGroupIcon className="h-5 w-5 text-gray-400" />
          <span className="font-medium">{team.name}</span>
        </div>
      ),
      searchValue: (team: Team) => `${team.id} ${team.name}`,
      sortValue: (team: Team) => team.name,
      sortable: true,
    },
    {
      key: 'createdAt',
      header: t('teams.createdAt'),
      accessor: (team: Team) => formatDate(team.createdAt, language),
      sortValue: (team: Team) => team.createdAt ? new Date(team.createdAt) : null,
      sortable: true,
    },
    {
      key: 'updatedAt',
      header: t('teams.updatedAt'),
      accessor: (team: Team) => formatDate(team.updatedAt, language),
      sortValue: (team: Team) => team.updatedAt ? new Date(team.updatedAt) : null,
      sortable: true,
    },
    {
      key: 'actions',
      header: t('teams.actions'),
      accessor: (team: Team) => (
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="ghost"
            icon={<UserPlusIcon className="h-4 w-4" />}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedTeam(team);
              setIsAddUserModalOpen(true);
            }}
          >
            {t('teams.addMember')}
          </Button>
          <Button
            size="sm"
            variant="danger"
            onClick={(e) => {
              e.stopPropagation();
              setSelectedTeam(team);
              setIsDeactivateModalOpen(true);
            }}
          >
            {t('teams.deactivateUsers')}
          </Button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            {t('teams.title')}
          </h1>
          <p className="mt-1 text-gray-600 dark:text-gray-400">
            Manage teams and their members
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="secondary"
            onClick={handleExport}
          >
            {t('common.export')}
          </Button>
          <Button
            variant="primary"
            icon={<PlusIcon className="h-5 w-5" />}
            onClick={() => setIsCreateModalOpen(true)}
          >
            {t('teams.createTeam')}
          </Button>
        </div>
      </div>

      {/* Teams Table */}
      <Table
        columns={columns}
        data={teamsData}
        loading={isLoading}
        searchable
        searchPlaceholder="Search teams..."
        selectable
        selectedRows={selectedRows}
        onSelectionChange={setSelectedRows}
        emptyMessage={t('teams.noTeams')}
        pagination
        pageSize={10}
        bulkActions={
          <>
            <Button
              size="sm"
              variant="danger"
              onClick={() => {
                const selectedTeamIds = Array.from(selectedRows).map(index => teamsData[index]?.id).filter(Boolean) as number[];
                if (selectedTeamIds.length > 0) {
                  if (confirm(`Are you sure you want to delete ${selectedTeamIds.length} team(s)?`)) {
                    deleteTeamsMutation.mutate(selectedTeamIds);
                  }
                }
              }}
              loading={deleteTeamsMutation.isPending}
            >
              {t('common.delete')}
            </Button>
          </>
        }
      />

      {/* Create Team Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setNewTeamName('');
        }}
        title={t('teams.createTeam')}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('teams.teamName')}
            </label>
            <input
              type="text"
              value={newTeamName}
              onChange={(e) => setNewTeamName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="Enter team name"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsCreateModalOpen(false);
                setNewTeamName('');
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => createTeamMutation.mutate(newTeamName)}
              loading={createTeamMutation.isPending}
              disabled={!newTeamName.trim()}
            >
              {t('common.save')}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Add User to Team Modal */}
      <Modal
        isOpen={isAddUserModalOpen}
        onClose={() => {
          setIsAddUserModalOpen(false);
          setSelectedUserId(null);
        }}
        title={`${t('teams.addMember')} - ${selectedTeam?.name}`}
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Select User
            </label>
            <select
              value={selectedUserId || ''}
              onChange={(e) => setSelectedUserId(Number(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">Select a user...</option>
              {usersData.map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name} ({user.username})
                </option>
              ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsAddUserModalOpen(false);
                setSelectedUserId(null);
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="primary"
              onClick={() => {
                if (selectedTeam && selectedUserId) {
                  addUserToTeamMutation.mutate({
                    teamId: selectedTeam.id,
                    userId: selectedUserId,
                  });
                }
              }}
              loading={addUserToTeamMutation.isPending}
              disabled={!selectedUserId}
            >
              Add User
            </Button>
          </div>
        </div>
      </Modal>

      {/* Deactivate Users Modal */}
      <Modal
        isOpen={isDeactivateModalOpen}
        onClose={() => {
          setIsDeactivateModalOpen(false);
          setUsersToDeactivate(new Set());
        }}
        title={`${t('teams.deactivateUsers')} - ${selectedTeam?.name}`}
        size="lg"
      >
        <div className="space-y-4">
          <p className="text-gray-600 dark:text-gray-400">
            {t('messages.confirmDeactivate')}
          </p>
          
          {/* User Selection List */}
          <div className="border border-gray-200 dark:border-gray-700 rounded-lg max-h-64 overflow-y-auto">
            {teamUsersData.length === 0 ? (
              <div className="p-4 text-center text-gray-500 dark:text-gray-400">
                No active users in this team
              </div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {teamUsersData.map((user) => (
                  <label
                    key={user.id}
                    className="flex items-center gap-3 p-3 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={usersToDeactivate.has(user.id)}
                      onChange={(e) => {
                        const newSet = new Set(usersToDeactivate);
                        if (e.target.checked) {
                          newSet.add(user.id);
                        } else {
                          newSet.delete(user.id);
                        }
                        setUsersToDeactivate(newSet);
                      }}
                      className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                    />
                    <div className="flex-1">
                      <div className="font-medium text-gray-900 dark:text-gray-100">
                        {user.name}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        @{user.username}
                      </div>
                    </div>
                  </label>
                ))}
              </div>
            )}
          </div>

          {usersToDeactivate.size > 0 && (
            <div className="text-sm text-gray-600 dark:text-gray-400 bg-yellow-50 dark:bg-yellow-900/20 p-3 rounded">
              ⚠️ {usersToDeactivate.size} user(s) will be deactivated and their open PRs will be automatically reassigned
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => {
                setIsDeactivateModalOpen(false);
                setUsersToDeactivate(new Set());
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="danger"
              onClick={() => {
                if (selectedTeam && usersToDeactivate.size > 0) {
                  deactivateUsersMutation.mutate({
                    teamId: selectedTeam.id,
                    userIds: Array.from(usersToDeactivate),
                  });
                }
              }}
              loading={deactivateUsersMutation.isPending}
              disabled={usersToDeactivate.size === 0}
            >
              {t('common.confirm')} ({usersToDeactivate.size})
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Teams;
