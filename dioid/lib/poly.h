#ifndef __POLY_H__
#define __POLY_H__

#define taille 64
#define taille_max_polynome 10000
#define LIMIT_MONOMES 1

void sort_gd(gd*, const int*);

class poly
{
    private:
        gd* data; // pointeur sur des objets monomes
        unsigned int n; // nombre d'��ents du polyn�e
        unsigned int nblock; // nombre de blocks de (taille*sizeof(gd)) allou� au polyn�e
        unsigned int simple; // vaut 1 lorsque le polyn�e est sous forme minimale

    public:
        poly(); // constructeur 0 : poly= (+00,-00) epsilon
        poly(const poly&); // constructuer 1 : serie initialisee avec un objet poly
        poly(const gd&);
        poly(long g, long d);
        poly(unsigned int, gd*); // constructeur 2 : initialise avec 1 vecteur de monome
        ~poly(); // destructeur

        poly& operator=(const poly&); // initialise avec un objet polynome
        poly& operator()(long g, long d); //initialisation avec deux monomes

        void init(unsigned int, gd*, int); // initialise avec 1 vecteur de monomes

        poly& operator=(const gd& gd1); // initialise avec un monome cast monome->polynome
        poly& init(long g, long d); // initialise 2 entiers cast monome->polynome

        void affecte(unsigned int, const gd*, unsigned int propre); // affectation avec 1 vecteur de monome
        gd&  getpol (int i) const {return data[i];} // pour acceder aux elements de data

        unsigned int getn() const {return n;}
        gd* getdata() {return data;}
        void popj(unsigned int j); // supprime l'��ent j
        void pop(); // supprime le dernier ��ent du polynome
        void add(const gd& m1); // ajoute un ��ent �la fin
        void simpli(); // tri et simplification du polynome
        void onlysimpli(); // simplification du polynome

        int operator==(const poly&);

        friend poly oplus(const poly&, const poly&); // somme de 2 polynomes, retourne un nouveau polynome
        friend poly oplus(const gd&, const gd&); // somme de deux monomes -> un polynome
        friend poly oplus(const poly& , const gd&); // somme d'1 polynome et d'un monome
        friend poly oplus(const gd&, const poly&); // somme d'un monome et d'1 polynome
        friend poly oplus(const poly&, const poly&, const poly&); // somme de 3 polynomes, retourne un nouveau polynome
        friend poly oplus(const poly&, const poly&, const poly&, const poly&); // somme de 4 polynomes, retourne un nouveau polynome

        friend poly otimes(const poly& poly1, const poly& poly2); // produit de polynome
        friend poly otimes(const poly& poly1, const gd& gd2); // produit d'un polynome par un monome
        friend poly otimes(const gd& gd1, const poly& poly2); // produit d'un monome par un polynome

        friend poly inf(const poly& poly1, const poly& poly2); // inf de polynome
        friend poly inf(const poly& poly1, const gd& gd2); // inf d'1 polynome et d'1 monome
        friend poly inf(const gd& gd1, const poly& poly2); // inf d'1 monome et d'1 polynome

        friend poly frac(const poly& poly1, const gd& gd2); // residuation d'1 polynome par 1 monome
        friend poly frac(const poly& poly1, const poly& poly2); //residuation de poly1/poly2
        friend poly frac(const gd& gd1, const poly& poly2); //residuation d'un monome par un polynome

        friend poly prcaus(poly&);
};

#endif

