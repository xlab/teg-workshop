#include "gd.h"
#include "poly.h"
#include <algorithm>

void sort_gd(gd* adtab, const int* n1)
{
    int j, sorted = 0;
    gd min;
    gd* prec;

    //std::vector<gd> data(*n1);
    // On regarde si le tableau n'est d��tri� auquel cas il faut
    // simplement appliquer les r�les de simplification
    prec = &(adtab[0]);
    sorted = 1;

    for (j = 1; j < (int)(*n1); ++j) { // on regarde si le tableau est deja tri�
        if (! ((*prec) < adtab[j])) {
            sorted = 0;
            j = *n1; // on sort de la boucle
        }

        prec = &(adtab[j]);
    }

    if (sorted == 0) { // on ne trie que s'il n'est pas deja trie
        std::sort(&(adtab[0]), &((adtab[*n1]))); // le dernier pointeur est sur l'�ement suivant le dernier !!!
        sorted = 1;
    }
}

void poly::affecte(unsigned int n1, const gd* vect, unsigned int propre)
// affectation avec 1 vecteur de monome
{
    unsigned int i, nbb;
    if (data != vect) {
        // si data==vect c'est lui meme
        nbb = (int) (n1 / taille) + 1;

        if (nbb > 10000 && LIMIT_MONOMES == 1) { //gestion d'une exception pour eviter de planter
            //std::cout<<"oh le polynome a "<<nbb*taille<<" monomes"<<endl;
            mem_limite l(0);
            throw (l);
        }

        if (nbb > nblock) {
            nblock = nbb;
            delete [] data;
            data = new gd[nblock * taille];
        }

        n = n1;
        if (n) {
            if (data) {
                for (i = 0; i < n; ++i) {
                    data[i] = vect[i];
                }

                simple = propre;
            } else {
                mem_limite l(1);
                throw (l);
            }
        } else {
            data = NULL;
        }
    }
}

poly::poly(void)
// constructeur 0 : poly= (+00,-00) epsilon
{
    n = 1;
    nblock = 1;
    data = new gd[taille];
    if (data == NULL) {
        mem_limite l(2);
        throw (l);
    }
    data->init(infty, _infty); // initialisation du monome
    simple = 1;
}

poly::poly(const poly& poly1)
// constructuer 1 : polynome initialise par un autre objet polynome
{
    n = 0; // par defaut on initialise le nombre d'elements a 0
    data = NULL; // et data ne pointe sur rien
    nblock = 0;
    affecte(poly1.n, poly1.data, poly1.simple);
}

poly::poly(unsigned int n1, gd* vect)
// constructeur 2 : initialise avec 1 vecteur de monome
{
    n = 0;
    data = NULL;
    nblock = 0;
    affecte(n1, vect, 0);
    simpli(); // on ne manipule que des polynomes croissants
}

poly::poly(const gd& gd1)
// constructeur 2 : initialise avec 1 vecteur de monome
{
    n = 1;
    nblock = 1;
    data = new gd[taille];
    if (data == NULL) {
        mem_limite l(3);
        throw (l);
    }
    data[0] = gd1;
    simple = 1;
}

poly::poly(long g, long d)
// constructeur 2 : initialise avec 2 entiers
{
    n = 1;
    nblock = 1;
    data = new gd[taille];
    if (data == NULL) {
        mem_limite l(4);
        throw (l);
    }
    data->init(g, d);
    simple = 1;
}

poly:: ~poly(void)
// destructeur
{
    if (n) {
        delete [] data;
    }
}

poly& poly::operator=(const poly& poly1)
// initialise avec un objet polynome
{
    if (&poly1 != this) {
        affecte(poly1.n, poly1.data, poly1.simple);
    }
    return *this;
}

void poly::init(unsigned int n1, gd* vect, int can)
// initialise avec 1 vecteur de monome
{
    //can=0;
    if (n1 < 1) { return; }// error }
    affecte(n1, vect, can);
    if (can == 0) { simpli(); } // on trie et on simplifie
    if (can == 2) { onlysimpli(); } // déjà trié on simplifie seulement
}

poly& poly::operator=(const gd& gd1)
// initialise avec un monome cast monome->polynome
{
    n = 1;
    affecte(n, &gd1, 0);
    return *this;
}

poly& poly::init (long g, long d) // initialise 2 entiers cast monome->polynome
{
    gd monome;
    n = 1;
    monome.init(g, d);
    affecte(n, &monome, 0);
    return *this;
}

poly& poly::operator ()(long g, long d) //initialisation avec deux entiers
{
    gd monome;

    monome(g, d);
    add(monome);
    simpli();
    return *this;
}

int poly::operator==(const poly& p1)
{
    unsigned int i = 0;

    if (p1.n != n) {
        return 0;
    } else {
        while (i < n) {
            if (data[i] != p1.data[i]) {
                return 0;
            }
            ++i;
        }
    }
    return 1;
}

void poly::simpli() // simplification du polynome
{
    gd monome;
    /*********** SIMPLIFICTAION DU POLYNOME poly *******************************/

    if (simple == 1) {
        return;
    }

    if (data) {
        sort_gd(data, (const int*) &n); // on trie suivant gamma, puis inverse de delta
        onlysimpli();
    }
}


void poly::onlysimpli() // simplification du polynome
{
    unsigned int i, j;
    gd monome;
    /*********** SIMPLIFICTAION DU POLYNOME poly *******************************/

    if (data) {
        i = 0;

        for (j = 1; j < n; ++j) {
            if (data[j].getd() > data[i].getd()) {
                ++i;
                data[i] = data[j];
            }
        }

        n = i + 1;

        if (data[0].getg() == _infty) { // on traite le cas ou le polynome commence par g=-00
            if (data[0].getd() != _infty) { // alors le polyn�e vaut Top
                monome.init(_infty, infty);
                affecte(1, &monome, 1);
                return;
            } else { //les premiers ��ent valent -oo,-oo <=> +oo,-oo
                while (data[0].getg() == _infty) {
                    popj(0);
                }
                // on le supprime si ce n'est pas le seul sinon il sera remplac�par epsilon
            }
        }

        //si le dernier element vaut epsilon, et qu'il y a plus d'un element on l'ote
        if (data[n - 1].getg() == infty && data[n - 1].getd() == _infty && n > 1) {
            popj(n - 1);
        }

        // si le dernier ��ent vaut +00,x <=> +00,-oo on l'ote
        if (data[n - 1].getg() == infty && data[n - 1].getd() != _infty) {
            popj(n - 1);
        }

        // si le premier ��ent vaut (x, -00) <=> +00,-00 on l'ote
        if (data[0].getd() == _infty && data[0].getg() != infty) {
            popj(0);
        }

        simple = 1;
    }
}


void poly::popj(unsigned int j)
{
    unsigned int i;

    if (data != NULL) {
        if (j < (n - 1)) {
            for (i = j; i < (n - 1); ++i) {
                data[i] = data[i + 1];
            } // si on ne supprime pas le dernier �ement
        }

        if (j < n) {
            pop();
        } // indice existant
    }
}

void poly::pop() // suppression du dernier ��ent du polynome s'il y en a au moins 2
{
    if (n > 1) {
        --n;
    } else {
        data->init(infty, _infty);
    }
}

void poly::add(const gd& m1) // ajout d'un ��ent �la fin d'un polyn�e
{
    gd* temp;
    unsigned int i;

    if (n >= (nblock * taille)) { // il faut refaire une affectation avec un block de plus
        temp = new gd[(nblock + 1) * taille]; // un tampon avec un block de plus

        if (temp == NULL) {
            // error
            return;
        }

        for (i = 0; i < n; ++i) {
            temp[i] = data[i];
        }

        temp[n] = m1;
        ++n;
        affecte(n, temp, 0); // un nouvelle affectation
        delete temp;
    } else {   // pas besoin de r�llouer
        if (n > 1) {
            data[n] = m1; // pas �se soucier du polyn�e nul
            ++n;
        } else {  // on regarde si le 1er vaut epsilon pour l'oter
            if ((data[0].getg() == infty) && (data[0].getd() == _infty)) {
                data[0] = m1;
            } else {
                data[n] = m1;
                ++n;
            }
        }
    }

    simple = 0;
}

poly oplus(const poly& poly1, const poly& poly2)
// somme de polynome, retour par valeur
// appel p=oplus(p1,p2); est possible car on a redefini l'operateur =
{
    gd* temp; // tableau temporaire
    gd* adlastgd = NULL;
    poly result;
    unsigned int nb = poly1.n + poly2.n;

    temp = new gd[nb];
    if (temp == NULL) {
        return result; // error
    }

    adlastgd = std::merge(&(poly1.data[0]), &(poly1.data[poly1.n]), &(poly2.data[0]), &(poly2.data[poly2.n]), temp);
    nb = adlastgd - temp;

    result.init(nb, temp, 2); // 2, pas besoin de trier juste simplification

    delete temp;
    return result;
}

poly oplus(const poly& poly1, const poly& poly2, const poly& poly3)
// somme de 3 polynomes, retour par valeur
// appel p=oplus(p1,p2,p3,p4); est possible car on a redefini l'operateur =
{
    gd* temp; // tableau temporaire
    gd* temp1;
    gd* adlastgd = NULL;
    gd* adlastgd1 = NULL;
    poly result;
    unsigned int nb = poly1.n + poly2.n;

    temp1 = new gd[nb];

    if (temp1 == NULL) {
        return result; // error
    }

    nb = nb + poly3.n;
    temp = new gd[nb];

    if (temp == NULL) {
        return result; // error
    }

    adlastgd1 = std::merge(&(poly1.data[0]), &(poly1.data[poly1.n]), &(poly2.data[0]), &(poly2.data[poly2.n]), temp1);
    adlastgd = std::merge(&(temp1[0]), adlastgd1, &(poly3.data[0]), &(poly3.data[poly3.n]), temp);
    nb = adlastgd - temp;

    result.init(nb, temp, 2);
    delete temp;
    delete temp1;
    return result;
}

poly oplus(const poly& poly1, const poly& poly2, const poly& poly3, const poly& poly4)
// somme de 4 polynomes, retour par valeur
// appel p=oplus(p1,p2,p3,p4); est possible car on a redefini l'operateur =
{
    gd* temp; // tableau temporaire
    gd* temp1;
    gd* temp2;
    gd* adlastgd = NULL;
    gd* adlastgd1 = NULL;
    gd* adlastgd2 = NULL;
    poly result;
    unsigned int nb1, nb, nb2;

    nb1 = poly1.n + poly2.n;
    temp1 = new gd[nb1];

    if (temp1 == NULL) {
        return result; // error
    }

    nb2 = poly4.n + poly3.n;
    temp2 = new gd[nb2];

    if (temp2 == NULL) {
        return result; // error
    }

    nb = nb1 + nb2;
    temp = new gd[nb];

    if (temp == NULL) {
        return result; // error
    }

    adlastgd1 = std::merge(&(poly1.data[0]), &(poly1.data[poly1.n]), &(poly2.data[0]), &(poly2.data[poly2.n]), temp1);
    adlastgd2 = std::merge(&(poly3.data[0]), &(poly3.data[poly3.n]), &(poly4.data[0]), &(poly4.data[poly4.n]), temp2);
    adlastgd = std::merge(&(temp1[0]), adlastgd1, &(temp2[0]), adlastgd2, temp);
    nb = adlastgd - temp;

    result.init(nb, temp, 2);
    delete temp;
    delete temp1;
    delete temp2;
    return result;
}

poly oplus(const poly& poly1, const gd& gd1)
// somme d'1 polynome avec un monome
{
    poly result(gd1);
    result = oplus(result, poly1);
    return result;
}

poly oplus(const gd& gd1, const poly& poly1)
// somme d'un monome avec 1 polynome
{
    poly result(gd1);
    result = oplus(result, poly1);
    return result;
}

poly oplus(const gd& gd1, const gd& gd2)
// somme de 2 monomes
{
    poly result(gd1);
    result = oplus(result, gd2);
    return result;
}

poly otimes(const poly& poly1, const poly& poly2)
// produit de polynome, retourne un nouveau polynome
// appel p=otimes(p1,p2); est possible car on a redefini l'operateur =
{
    poly result;
    const poly* p1, *p2, *ptemp;
    gd** temp = NULL; // tableau temporaire de monomes
    gd** tampon = NULL; // tableau temporaire de monomes
    gd* adlastgd;
    int* tabtaille = NULL; // tableau des tailles des tableaux de monomes
    unsigned int i, j, k, nbpoly;
    unsigned int taillediv2 = 0;
    poly poly1temp, poly2temp;

    // s'ils ne sont pas sous forme simple : en raison de méthode Add
    p1 = &poly1;
    p2 = &poly2;

    if (poly1.simple == 0) {
        poly1temp = poly1;
        poly1temp.simpli();
        p1 = &poly1temp;
    }

    if (poly2.simple == 0) {
        poly2temp = poly2;
        poly2temp.simpli();
        p2 = &poly2temp;
    }

    if (!(p1->getn() < p2->getn())) {
        ptemp = p1;
        p1 = p2;
        p2 = ptemp;
    }

    temp = new gd * [p1->getn()]; // tableau de polyn�es

    if (temp == NULL) {
        return result; // error
    }

    taillediv2 = (int)((p1->getn() + 1) / 2);

    tampon = new gd * [taillediv2]; // tableau de poyn�es apr� 1ere fusion /2 +1

    if (tampon == NULL) {
        return result; // error
    }

    tabtaille = new int [p1->getn()]; // tableau de int contenant la taille des polyn�es

    if (tabtaille == NULL) {
        return result; // error
    }

    for (j = 0; j < p1->getn(); ++j) {
        temp[j] = new gd [p2->getn()]; // creation des tableaux de monomes
        if (temp[j] == NULL) {
            return result; // error
        }
        tabtaille[j] = p2->getn(); // leur taille sera celle de p2
    }

    for (j = 0; j < taillediv2; ++j) {
        tampon[j] = new gd [p2->getn() * 2]; // double de taille pour la fusion de 2 tableaux
        if (tampon[j] == NULL) {
            return result; // error
        }
    }

    // on fait le produit des mon�es de p1 par le polynome p2
    for (j = 0; j < p1->getn(); ++j) {
        for (k = 0; k < p2->getn(); ++k) {
            temp[j][k] = otimes(p1->getpol(j), p2->getpol(k));
        }
    }

    nbpoly = p1->getn(); // initialmeent on a donc nbpoly

    while (nbpoly > 1) {
        i = 0;
        j = 0;

        // 1ere �ape de fusion
        while (i < nbpoly) {
            k = i + 1;
            if (k < nbpoly) { // s'il en reste bien 2 �fusionner
                // fusion
                adlastgd = std::merge(&temp[i][0], &temp[i][tabtaille[i]], &temp[k][0], &temp[k][tabtaille[k]], &tampon[j][0]);
                tabtaille[j] = adlastgd - &tampon[j][0];

                delete [] temp[i];
                delete [] temp[k];
            } else {
                tabtaille[j] = tabtaille[i];
                tampon[j] = temp[i];
            }
            i = i + 2;
            ++j;
        }

        // on pr�are pour une nouvelle �ape de fusion
        nbpoly = j;
        taillediv2 = (int)((nbpoly + 1) / 2);

        for (j = 0; j < nbpoly; ++j) {
            temp[j] = tampon[j]; // on permute les pointeurs pour r�up�er les tableaux fusionn�
        }

        for (j = 0, k = 0; j < taillediv2; ++j) { // on alloue de nouveaux tableaux pour les futurs fusionn�
            tampon[j] = NULL;

            if (k + 1 < nbpoly) {
                tampon[j] = new gd[tabtaille[k] + tabtaille[k + 1]];
            } // si fusion somme des tailles

            else {
                tampon[j] = new gd[tabtaille[k]];
            } // si nombre impair, pas de fusion pour le dernier

            if (tampon[j] == NULL) {
                return result; // error
            }

            k = k + 2;
        }
    }

    result.init(tabtaille[0], temp[0], 2);

    delete temp[0];
    delete temp;
    delete tabtaille;
    delete tampon[0];
    delete tampon;

    return result;
}

poly otimes(const poly& poly1, const gd& gd2) // produit d'un polynome par un monome
{
    poly poly2;
    poly2 = gd2; //cast gd->poly
    return otimes(poly1, poly2);
}

poly otimes(const gd& gd1, const poly& poly2) // produit d'un monome par un polynome
{
    poly poly1;
    poly1 = gd1; //cast gd->poly
    return otimes(poly1, poly2);
}

poly inf(const poly& poly1, const poly& poly2)
// inf de polynome, retourne un nouveau polynome
{
    int M;
    gd gdtemp;
    poly result;
    const poly* p1, *p2, *ptemp;
    gd** temp = NULL; // tableau temporaire de monomes
    gd** tampon = NULL; // tableau temporaire de monomes
    gd* adlastgd;
    int* tabtaille = NULL; // tableau des tailles des tableaux de monomes
    unsigned int i, j, k, nbpoly;
    unsigned int taillediv2 = 0;
    poly poly1temp, poly2temp;

    p1 = &poly1;
    p2 = &poly2;

    // s'ils ne sont pas sous forme simple : en raison de méthode Add
    if (poly1.simple == 0) {
        poly1temp = poly1;
        poly1temp.simpli();
        p1 = &poly1temp;
    }

    if (poly2.simple == 0) {
        poly2temp = poly2;
        poly2temp.simpli();
        p2 = &poly2temp;
    }

    // p1 est le plus petit polynomes
    if (!(p1->getn() < p2->getn())) {
        ptemp = p1;
        p1 = p2;
        p2 = ptemp;
    }

    temp = new gd * [p1->getn()]; // tableau de polyn�es

    if (temp == NULL) {
        return result; // error
    }

    taillediv2 = (int)((p1->getn() + 1) / 2);

    tampon = new gd * [taillediv2]; // tableau de poyn�es apr� 1ere fusion /2 +1

    if (tampon == NULL) {
        return result; // error
    }

    tabtaille = new int [p1->getn()]; // tableau de int contenant la taille des polyn�es

    if (tabtaille == NULL) {
        return result; // error
    }


    for (j = 0; j < p1->getn(); ++j) {
        temp[j] = new gd [p2->getn()]; // creation des tableaux de monomes
        if (temp[j] == NULL) {
            return result; // error
        }
        tabtaille[j] = p2->getn(); // leur taille sera celle de p2 par d�aut
    }

    for (j = 0; j < taillediv2; ++j) {
        tampon[j] = new gd [p2->getn() * 2]; // double de taille pour la fusion de 2 tableaux
        if (tampon[j] == NULL) {
            return result; // error
        }
    }

    // on fait l'inf des mon�es de p1 par le polynome p2
    for (j = 0; j < p1->getn(); ++j) {
        temp[j][0] = inf(p1->getpol(j), p2->getpol(0));

        for (k = 1, M = 1; k < p2->getn(); ++k) {
            temp[j][M] = inf(p1->getpol(j), p2->getpol(k));

            if (temp[j][M].getd() == temp[j][M - 1].getd()) { // on bouge plus en delta c'est fini
                tabtaille[j] = M;
                k = p2->getn();
            } else {
                if (temp[j][M].getg() == temp[j][M - 1].getg()) { // il ne faut garder que le dernier
                    temp[j][M - 1] = temp[j][M];
                    tabtaille[j] = tabtaille[j] - 1; // on ote un ��ent
                } else {
                    ++M;
                }
            }

        }
    }

    nbpoly = p1->getn(); // initialmeent on a donc nbpoly

    while (nbpoly > 1) {
        i = 0;
        j = 0;

        // 1ere �ape de fusion
        while (i < nbpoly) {
            k = i + 1;
            if (k < nbpoly) { // s'il en reste bien 2 �fusionner
                // fusion
                adlastgd = std::merge(&temp[i][0], &temp[i][tabtaille[i]], &temp[k][0], &temp[k][tabtaille[k]], &tampon[j][0]);
                tabtaille[j] = adlastgd - &tampon[j][0];

                delete [] temp[i];
                delete [] temp[k];
            } else {
                tabtaille[j] = tabtaille[i];
                tampon[j] = temp[i];
            }
            i = i + 2;
            ++j;
        }

        // on pr�are pour une nouvelle �ape de fusion
        nbpoly = j;
        taillediv2 = (int)((nbpoly + 1) / 2);

        for (j = 0; j < nbpoly; ++j) {
            temp[j] = tampon[j]; // on permute les pointeurs pour r�up�er les tableaux fusionn�
        }

        for (j = 0, k = 0; j < taillediv2; ++j) { // on alloue de nouveaux tableaux pour les futurs fusionn�
            tampon[j] = NULL;

            if (k + 1 < nbpoly) {
                tampon[j] = new gd[tabtaille[k] + tabtaille[k + 1]];
            } // si fusion somme des tailles

            else {
                tampon[j] = new gd[tabtaille[k]];
            } // si nombre impair, pas de fusion pour le dernier

            if (tampon[j] == NULL) {
                return result; // error
            }

            k = k + 2;
        }
    }

    result.init(tabtaille[0], temp[0], 2);

    delete temp[0];
    delete temp;
    delete tabtaille;
    delete tampon[0];
    delete tampon;

    return result;
}

poly inf(const poly& poly1, const gd& gd2)
// inf d'1 polynome et d'1 monome
{
    return inf(gd2, poly1);
}

poly inf(const gd& gd1, const poly& poly2)
// inf d'1 monome et d'1 polynome
{
    poly result;
    gd* temp;
    unsigned int i, j, M;

    if (!poly2.simple) {
        return result; // error
    }

    temp = new gd[poly2.getn()]; // tableau de mon�es
    j = poly2.getn();

    if (temp == NULL) {
        return result; // error
    }

    temp[0] = inf(gd1, poly2.getpol(0));

    for (i = 1, M = 1; i < poly2.getn(); ++i) {
        temp[M] = inf(gd1, poly2.getpol(i));

        if (temp[M].getd() == temp[M - 1].getd()) { // on bouge plus en delta c'est fini
            j = M;
            i = poly2.getn();
        }

        if (temp[M].getg() == temp[M - 1].getg()) { // il ne faut garder que le dernier
            temp[M - 1] = temp[M];
            --j; // on ote un ��ent
        } else {
            ++M;
        }
    }

    result.init(j, temp, 1);
    delete temp;
    return result;
}

poly frac(const poly& poly1, const gd& gd2)
// residuation d'1 polynome par un monome, retourne un nouveau polynome
{
    poly result;
    gd* temp;
    unsigned int k;

    if (!poly1.simple) {
        return result; // error
    }

    temp = new gd [poly1.getn()]; // tableau de mon�es

    if (temp == NULL) {
        return result; // error
    }

    temp[0] = frac(poly1.getpol(0), gd2);

    for (k = 1; k < poly1.getn(); ++k) {
        temp[k] = frac(poly1.getpol(k), gd2);
    }

    result.affecte(poly1.getn(), temp, 2); // d��tri� juste simplifier
    return result;
}

poly frac(const poly& poly1, const poly& poly2)
// residuation d'1 polynome par 1 polynome, retourne un nouveau polynome
{
    poly result, tampon;
    unsigned int j;

    result = frac(poly1, poly2.data[0]);

    for (j = 1; j < poly2.n; ++j) {
        tampon = frac(poly1, poly2.data[j]);
        result = inf(result, tampon);
    }

    return result;
}

poly frac(gd& gd1, poly& poly2)
// residuation d'1 monome par 1 polynome, retourne un nouveau polynome
// appel p=frac(m1,p2); est possible car on a redefini l'operateur =
{
    poly poly1, result;
    poly1 = gd1; //cast gd->poly
    result = frac(poly1, poly2);
    return result;
}

poly prcaus(poly& p)
{
    poly local;
    int i = (p.getn() - 1);

    while (p.getpol(i).getd() >= 0 && p.getpol(i).getg() >= 0 && i >= 0) {
        local.add(p.getpol(i));
        --i;
    }

    // si i==-1 tous les monomes de p etaient causaux
    // sinon il y a peut �re la partie causale du i-1 ieme �ajouter
    if (i >= 0 && p.getpol(i).getd() >= 0) {
        local.add(gd(0, p.getpol(i).getd()));
    }

    local.simpli();
    return local;
}
