# rclone 설정 단계별 가이드

## 현재 상태: `~/bin/rclone config` 실행 완료

---

## 화면에 보이는 내용과 입력할 내용:

### 1단계: New remote 만들기

```
No remotes found, make a new one?
n) New remote
s) Set configuration password
q) Quit config
n/s/q>
```

**입력:** `n` 엔터

---

### 2단계: Remote 이름 정하기

```
Enter name for new remote.
name>
```

**입력:** `gdrive` 엔터

---

### 3단계: Storage 타입 선택

```
Option Storage.
Type of storage to configure.
Choose a number from below, or type in your own value.
 1 / 1Fichier
   \ (fichier)
 2 / Akamai NetStorage
   \ (netstorage)
...
17 / Google Drive
   \ (drive)
...
Storage>
```

**입력:** `drive` 엔터
(또는 숫자가 보이면 Google Drive의 번호 입력, 보통 17번)

---

### 4단계: Google Application Client Id

```
Option client_id.
Google Application Client Id
Setting your own is recommended.
See https://rclone.org/drive/#making-your-own-client-id for how to create your own.
If you leave this blank, it will use an internal key which is low performance.
Enter a value. Press Enter to leave empty.
client_id>
```

**입력:** 그냥 엔터 (비워둠)

---

### 5단계: OAuth Client Secret

```
Option client_secret.
OAuth Client Secret.
Leave blank normally.
Enter a value. Press Enter to leave empty.
client_secret>
```

**입력:** 그냥 엔터 (비워둠)

---

### 6단계: Scope 선택 (권한 범위)

```
Option scope.
Scope that rclone should use when requesting access from drive.
Choose a number from below, or type in your own value.
Press Enter to leave empty.
 1 / Full access all files, excluding Application Data Folder.
   \ (drive)
 2 / Read-only access to file metadata and file contents.
   \ (drive.readonly)
...
scope>
```

**입력:** `1` 엔터

---

### 7단계: Root folder ID

```
Option root_folder_id.
ID of the root folder.
Leave blank normally.
...
Enter a value. Press Enter to leave empty.
root_folder_id>
```

**입력:** 그냥 엔터 (비워둠)

---

### 8단계: Service Account Credentials JSON

```
Option service_account_file.
Service Account Credentials JSON file path.
Leave blank normally.
...
Enter a value. Press Enter to leave empty.
service_account_file>
```

**입력:** 그냥 엔터 (비워둠)

---

### 9단계: Advanced Config

```
Edit advanced config?
y) Yes
n) No (default)
y/n>
```

**입력:** `n` 엔터

---

### 10단계: Auto Config (중요!)

```
Use auto config?
 * Say Y if not sure
 * Say N if you are working on a remote or headless machine

y) Yes (default)
n) No
y/n>
```

**입력:** `y` 엔터

→ **이 시점에서 브라우저가 자동으로 열립니다**

---

## 🌐 브라우저 인증 단계

1. 브라우저가 열리면 **Google 계정 선택**
2. **"rclone이 Google 계정에 액세스하려고 합니다"** 화면
3. **"허용"** 버튼 클릭
4. **"Success! All done. Please go back to rclone."** 메시지 확인
5. 브라우저 탭 닫고 터미널로 돌아오기

---

### 11단계: Shared Drive (Team Drive)

```
Configure this as a Shared Drive (Team Drive)?
y) Yes
n) No (default)
y/n>
```

**입력:** `n` 엔터

---

### 12단계: 설정 확인

```
Configuration complete.
Options:
- type: drive
- scope: drive
...
Keep this "gdrive" remote?
y) Yes this is OK (default)
e) Edit this remote
d) Delete this remote
y/e/d>
```

**입력:** `y` 엔터

---

### 13단계: 종료

```
Current remotes:

Name                 Type
====                 ====
gdrive               drive

e) Edit existing remote
n) New remote
d) Delete remote
r) Rename remote
c) Copy remote
s) Set configuration password
q) Quit config
e/n/d/r/c/s/q>
```

**입력:** `q` 엔터

---

## ✅ 설정 완료!

이제 테스트:

```bash
~/bin/rclone lsd gdrive:
```

Google Drive의 폴더 목록이 보이면 성공입니다!

---

## 요약: 입력 순서

1. `n` (new)
2. `gdrive` (이름)
3. `drive` (storage)
4. (엔터 - client_id)
5. (엔터 - client_secret)
6. `1` (full access)
7. (엔터 - root_folder_id)
8. (엔터 - service_account)
9. `n` (advanced config 안 함)
10. `y` (auto config 사용) → **브라우저 인증**
11. `n` (shared drive 아님)
12. `y` (설정 확인)
13. `q` (종료)
