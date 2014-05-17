#include "tools.h"

int gcd(int a, int b)
{
    int r;

    while (b > 0) {
        r = a % b;
        a = b;
        b = r;
    }

    return a;
}

int lcm(int a, int b)
{
    int a_sauve, b_sauve;
    a_sauve = a;
    b_sauve = b;
    a = gcd(a, b);
    
    return (a_sauve * b_sauve) / a;
}

