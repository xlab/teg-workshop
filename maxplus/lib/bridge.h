#ifndef BRIDGE_H
#define BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

void* newGd(int g, int d);
void* newPoly();
void* newSerie(void* p, void* q, void* r);

void freeGd(void* p);
void freePoly(void* p);
void freeSerie(void* s);

void appendPoly(void* p, int g, int d);
void simplyPoly(void* p);
unsigned int lenPoly(void* p);
void* getPoly(void* p, int n);

int getG(void* m);
int getD(void* m);
void* getP(void* s);
void* getQ(void* s);
void* getR(void* s);

void* starPoly(void* p);
void* starSerie(void* s);
void* oplusPoly(void* p1, void* p2);
void* otimesPoly(void* p1, void* p2);
void* oplusSerie(void* s1, void* s2);
void* otimesSerie(void* s1, void* s2);

#ifdef __cplusplus
}
#endif

#endif