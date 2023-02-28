import { AbstractControl, ValidationErrors, ValidatorFn } from "@angular/forms";

    export function PasswordStrenghtValidator(): ValidatorFn {
        return (control: AbstractControl) : ValidationErrors | null => {
          const value = control.value;
      
          if (!value) {
            return null;
          }
      
          const hasUpperCase = /[A-Z]+/.test(value)
      
          const hasLowerCase = /[a-z]+/.test(value);
      
          const hasNumeric = /[0-9]+/.test(value);
      
          const passwordValid = hasUpperCase && hasLowerCase && hasNumeric;
      
          return !passwordValid ? {passwordStrength:true}: null;
        }
      }

      export function PasswordSpecialCharacterValidator() {
        return (control: AbstractControl) : ValidationErrors | null => {
            const value = control.value;

            if (!value) {
                return null;
            }

            const noSpecialCharacter = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(value);

            const passwordValid = noSpecialCharacter;

            return passwordValid ? {noSpecialCharacter: true}: null;
        } 
    }
