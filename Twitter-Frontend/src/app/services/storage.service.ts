import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class StorageService {

  constructor() { }

  getRoleFromToken(): string {
    const authTokenToken = window.localStorage.getItem('authToken')
    if (authTokenToken) {
      const tokenSplit = authTokenToken.split('.')
      const decoded = decodeURIComponent(encodeURIComponent(window.atob(tokenSplit[1])))
      const obj = JSON.parse(decoded)
      return obj["userType"]
    }
    return ""
  }

  getUsernameFromToken(): string {
    const authTokenToken = window.localStorage.getItem('authToken')
    if (authTokenToken) {
      const tokenSplit = authTokenToken.split('.')
      const decoded = decodeURIComponent(encodeURIComponent(window.atob(tokenSplit[1])))
      const obj = JSON.parse(decoded)
      return obj["cognito:username"]
    }
    return ""
  }

  decodedToken(): any {
    const authTokenToken = window.localStorage.getItem('authToken')
    if (authTokenToken) {
      const tokenSplit = authTokenToken.split('.')
      const decoded = decodeURIComponent(encodeURIComponent(window.atob(tokenSplit[1])))
      const obj = JSON.parse(decoded)
      console.log(JSON.stringify(obj))
      return obj
    }
    return null
  }

}
