declare namespace User {
  type User = {
    id: string;
    username: string;
    role: string;
  };

  type GetUsersRequest = {
    page: number;
    limit: number;
  }

  type GetUsersResponse = {
    page: number;
    limit: number;
    totalPages: number;
    totalUsers: number;
    data: User[];
  }

  type CreateUserRequest = {
    name: string;
    password: string;
    role: string;
  }

  type DeleteUserRequest = {
    id: string;
  }
}
