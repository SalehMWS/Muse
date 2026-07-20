/**
 * Settings owns its own profile shape rather than importing features/auth,
 * since cross-feature imports are not allowed. Both read the same endpoint.
 */
export interface ProfileDto {
  id: string;
  email: string;
  display_name: string;
  status: string;
  email_verified: boolean;
  created_at: string;
}

export interface Profile {
  id: string;
  email: string;
  displayName: string;
  status: string;
  emailVerified: boolean;
  createdAt: string;
}

export function toProfile(dto: ProfileDto): Profile {
  return {
    id: dto.id,
    email: dto.email,
    displayName: dto.display_name,
    status: dto.status,
    emailVerified: dto.email_verified,
    createdAt: dto.created_at,
  };
}
