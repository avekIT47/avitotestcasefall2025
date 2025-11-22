// Типы данных для работы с API

export interface Team {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: number;
  username: string;
  name: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  teams?: Team[];
}

export interface PullRequest {
  id: number;
  title: string;
  author?: User;
  authorId: number;
  reviewers?: User[];
  team?: Team;
  teamId: number;
  status: 'open' | 'merged' | 'closed';
  createdAt: string;
  updatedAt: string;
  mergedAt?: string;
}

export interface Statistics {
  totalPRs: number;
  openPRs: number;
  mergedPRs: number;
  userStats: UserStatistic[];
  teamStats: TeamStatistic[];
}

export interface UserStatistic {
  userId: number;
  userName: string;
  assignmentCount: number;
}

export interface TeamStatistic {
  teamId: number;
  teamName: string;
  prCount: number;
}

export interface CreateTeamRequest {
  name: string;
}

export interface CreateUserRequest {
  username: string;
  name: string;
  isActive?: boolean;
}

export interface UpdateUserRequest {
  name?: string;
  isActive?: boolean;
}

export interface CreatePullRequestRequest {
  title: string;
  authorId: number;
  teamId: number;
}

export interface AddUserToTeamRequest {
  userId: number;
}

export interface RemoveUserFromTeamRequest {
  userId: number;
}

export interface DeactivateUsersRequest {
  userIds: number[];
}

export interface ReassignReviewerRequest {
  oldReviewerId: number;
  newReviewerId?: number;
}

export interface ApiError {
  error: string;
  message?: string;
}
