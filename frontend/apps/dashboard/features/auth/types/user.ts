export interface User {
  id: string;
  email: string;
  displayName: string;
  status: string;
  emailVerified: boolean;
  createdAt: string;
}

export interface UserDto {
  id: string;
  email: string;
  display_name: string;
  status: string;
  email_verified: boolean;
  created_at: string;
}

export function toUser(dto: UserDto): User {
  return {
    id: dto.id,
    email: dto.email,
    displayName: dto.display_name,
    status: dto.status,
    emailVerified: dto.email_verified,
    createdAt: dto.created_at,
  };
}
