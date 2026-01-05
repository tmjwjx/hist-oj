#include <bits/stdc++.h>
#ifdef LOCAL
    #include<pprint.hpp>
#else
    #define debug(...)
#endif
using namespace std;
using i64=long long;
unsigned seed=chrono::system_clock::now().time_since_epoch().count();
mt19937 rng(seed);
i64 rand(i64 l,i64 r){return rng()%(r-l+1)+l;}
template<class T>
bool chmax(T &ans,T t){
    if(ans>t) return false;
    ans=t;
    return true;
}
template<class T>
bool chmin(T &ans,T t){
    if(ans<=t) return false;
    ans=t;
    return true;
}
inline i64 read() {
    int x = 0, f = 1; 
    char ch = getchar();
    while (ch < '0' || ch > '9') {
        if (ch == '-') f = -1;
        ch = getchar();
    }
    while (ch >= '0' && ch <= '9') {
        x = x * 10 + (ch - '0'); 
        ch = getchar();
    }
    return x * f;
}
inline void write(long long x) {
    if (x < 0) {
        putchar('-');
        x = -x;
    }
    if (x > 9) {
        write(x / 10);
    }
    putchar(x % 10 + '0');
}
void addw(vector<vector<pair<int,int>>>&e,int u,int v,int w){
    e[u].push_back({v,w});
    e[v].push_back({u,w});
}
void add(vector<vector<int>>&e,int u,int v){
    e[u].push_back(v);
    e[v].push_back(u);
}
void solve(int task){
   int a,b;
   cin>>a>>b;
   cout<<a+b;
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);cout.tie(nullptr);
#ifdef LOCAL
    freopen("Bingbong.in","r",stdin);
#endif
   int t=1;
   for(int i=1;i<=t;i++){
    solve(i);
   }
  return 0;
}
