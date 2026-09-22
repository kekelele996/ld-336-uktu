import { Injectable, computed, signal } from '@angular/core';
import { LoginResp, User } from '../models';

const TOKEN_KEY = 'medasset_token';
const USER_KEY = 'medasset_user';

function loadUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

@Injectable({ providedIn: 'root' })
export class AuthStore {
  private tokenSig = signal<string>(localStorage.getItem(TOKEN_KEY) ?? '');
  private userSig = signal<User | null>(loadUser());

  readonly token = this.tokenSig.asReadonly();
  readonly user = this.userSig.asReadonly();
  readonly isLoggedIn = computed(() => !!this.tokenSig());

  setAuth(resp: LoginResp): void {
    this.tokenSig.set(resp.token);
    this.userSig.set(resp.user);
    localStorage.setItem(TOKEN_KEY, resp.token);
    localStorage.setItem(USER_KEY, JSON.stringify(resp.user));
  }

  setUser(user: User): void {
    this.userSig.set(user);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }

  clear(): void {
    this.tokenSig.set('');
    this.userSig.set(null);
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }

  hasRole(...roles: string[]): boolean {
    const u = this.userSig();
    return !!u && roles.includes(u.role);
  }
}
