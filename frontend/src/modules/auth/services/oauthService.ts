import api from '@/core/api/client';
import type { 
  OAuthAuthURLResponse, 
  OAuthStatusResponse, 
  OAuthDisconnectResponse,
  OAuthProvider 
} from '@/types';

export const OAuthService = {
  /**
   * Get OAuth authorization URL for connecting a provider
   * @param provider - OAuth provider (e.g., 'google')
   * @param scopes - Array of scopes to request (defaults to calendar scopes)
   * @returns Authorization URL to redirect user to
   */
  async getAuthURL(provider: OAuthProvider = 'google', scopes?: string[]): Promise<string> {
    const params = new URLSearchParams();
    params.append('provider', provider);
    if (scopes && scopes.length > 0) {
      params.append('scopes', scopes.join(','));
    }
    
    const response = await api.get<OAuthAuthURLResponse>(`/api/oauth/auth-url?${params.toString()}`);
    return response.data.auth_url;
  },

  /**
   * Get OAuth connection status for the current user
   * @param provider - OAuth provider to check (optional, defaults to google)
   * @returns Connection status and list of all connections
   */
  async getStatus(provider?: OAuthProvider): Promise<OAuthStatusResponse> {
    const params = new URLSearchParams();
    if (provider) {
      params.append('provider', provider);
    }
    
    const response = await api.get<OAuthStatusResponse>(`/api/oauth/status?${params.toString()}`);
    return response.data;
  },

  /**
   * Disconnect an OAuth provider
   * @param provider - OAuth provider to disconnect
   * @returns Success response
   */
  async disconnect(provider: OAuthProvider = 'google'): Promise<OAuthDisconnectResponse> {
    const params = new URLSearchParams();
    params.append('provider', provider);
    
    const response = await api.post<OAuthDisconnectResponse>(`/api/oauth/disconnect?${params.toString()}`);
    return response.data;
  },

  /**
   * Open OAuth connection in a popup window
   * @param authUrl - The authorization URL from getAuthURL
   * @returns Promise that resolves when popup closes
   */
  openOAuthPopup(authUrl: string): Promise<{ success: boolean; error?: string }> {
    return new Promise((resolve) => {
      // Calculate popup position for center of screen
      const width = 500;
      const height = 600;
      const left = window.screenX + (window.outerWidth - width) / 2;
      const top = window.screenY + (window.outerHeight - height) / 2;

      // Open popup
      const popup = window.open(
        authUrl,
        'oauth-popup',
        `width=${width},height=${height},left=${left},top=${top},popup=1`
      );

      if (!popup) {
        resolve({ success: false, error: 'Failed to open popup. Please allow popups for this site.' });
        return;
      }

      // Listen for messages from popup
      const handleMessage = (event: MessageEvent) => {
        // Verify origin - in production, check against your backend URL
        if (event.data && event.data.type === 'oauth-callback') {
          window.removeEventListener('message', handleMessage);
          clearInterval(checkClosed);
          
          if (event.data.success) {
            resolve({ success: true });
          } else {
            resolve({ success: false, error: event.data.error || 'OAuth failed' });
          }
        }
      };

      window.addEventListener('message', handleMessage);

      // Check if popup is closed manually
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
          window.removeEventListener('message', handleMessage);
          resolve({ success: false, error: 'Popup closed before completion' });
        }
      }, 500);
    });
  },

  /**
   * Check if user has Google Calendar connected
   * @returns true if connected with calendar scope
   */
  async isCalendarConnected(): Promise<boolean> {
    try {
      const status = await this.getStatus('google');
      // Handle null connections array
      const connections = status.connections ?? [];
      const googleConnection = connections.find(
        conn => conn.provider === 'google' && conn.is_connected
      );
      
      if (!googleConnection) return false;
      
      // Check for calendar scopes
      const hasCalendarScope = googleConnection.scopes?.some(scope => 
        scope.includes('calendar')
      ) ?? false;
      
      return hasCalendarScope;
    } catch {
      return false;
    }
  }
};
