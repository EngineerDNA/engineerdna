import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/client';
import type { WidgetProps } from './widget-registry';

export function TeamMemberCardsWidget({ config, onOpenModal }: WidgetProps) {
  const teamId = config.team_id as string | undefined;
  const sortBy = (config.sort_by as string) || 'name';

  const { data: membersData, isLoading } = useQuery({
    queryKey: ['team-members', teamId],
    queryFn: () => api.getTeamMembers(teamId || ''),
    enabled: !!teamId,
    staleTime: 5 * 60 * 1000,
  });

  const members = membersData?.members || [];

  const sortedMembers = [...members].sort((a, b) => {
    if (sortBy === 'name') {
      return a.name.localeCompare(b.name);
    }
    return 0;
  });

  if (!teamId) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
          Configure team_id to view team members
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full">
        <div className="animate-pulse grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="h-32 bg-gray-300 dark:bg-gray-700 rounded" />
          ))}
        </div>
      </div>
    );
  }

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">Team Members</h3>
        <span className="text-xs text-gray-500 dark:text-gray-400">{members.length} members</span>
      </div>

      <div className="flex-1 overflow-auto">
        {members.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-gray-500 dark:text-gray-400">
            No team members found
          </div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {sortedMembers.map((member) => (
              <div
                key={member.id}
                onClick={() => onOpenModal?.('engineer-detail', { engineerId: member.id })}
                className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 cursor-pointer hover:shadow-lg hover:border-blue-500 dark:hover:border-blue-400 transition-all"
              >
                <div className="flex flex-col items-center text-center">
                  <div className="w-16 h-16 rounded-full bg-blue-600 dark:bg-blue-500 flex items-center justify-center text-white font-bold text-lg mb-3">
                    {getInitials(member.name)}
                  </div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate w-full mb-1">
                    {member.name}
                  </p>
                  {member.role_id && (
                    <p className="text-xs text-gray-500 dark:text-gray-400 truncate w-full">
                      {member.role_id}
                    </p>
                  )}
                  <div className="mt-2 w-full">
                    <div className="flex items-center justify-center gap-1">
                      <div
                        className={`w-2 h-2 rounded-full ${member.active ? 'bg-green-500' : 'bg-gray-400'}`}
                      />
                      <span className="text-xs text-gray-600 dark:text-gray-400">
                        {member.active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
