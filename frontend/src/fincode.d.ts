import type { FincodeInstance } from '@fincode/js';

declare global {
  interface Window {
    Fincode: (publicKey: string) => FincodeInstance;
  }
}
